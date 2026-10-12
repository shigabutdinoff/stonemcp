package mindconnect

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

// ErrPacketRejected означает, что игра не разобрала пакет и ответила prt_err.
var ErrPacketRejected = errors.New("mindconnect: игра отвергла пакет (prt_err)")

// ErrMalformedResponse означает ответ, который нельзя разобрать.
var ErrMalformedResponse = errors.New("mindconnect: ответ искажён")

// ErrTruncatedResponse означает ответ, оборванный до сторожевого значения.
var ErrTruncatedResponse = errors.New("mindconnect: ответ оборван")

// ErrIncompleteResponse означает, что ответ ещё не дочитан.
var ErrIncompleteResponse = errors.New("mindconnect: ответ не дочитан")

// ResponseReader собирает один ответ игры из нескольких чтений.
type ResponseReader struct {
	id      int
	want    int
	buf     []byte
	total   int
	seps    int
	tailPos int
	done    bool
	err     error
	values  []Value
}

// NewResponseReader ждёт ответ на пакет id из want значений.
// При want меньше 1 или id вне 1..MaxID читатель сразу несёт липкую ошибку.
func NewResponseReader(id, want int) *ResponseReader {
	if want < 1 || id < 1 || id > MaxID {
		err := fmt.Errorf("%w: ответ на id %d из %d значений", ErrInvalidRequest, id, want)
		return &ResponseReader{id: id, want: want, done: true, err: err}
	}
	return &ResponseReader{id: id, want: want, values: make([]Value, 0, want)}
}

// Feed добавляет байты ответа и сообщает, дочитан ли он; ошибка липкая.
// Управляющий байт, чужой id, лишнее значение и ответ длиннее MaxResponse
// дают ErrMalformedResponse; отказ prt_err разрешает только Stalled.
func (r *ResponseReader) Feed(p []byte) (bool, error) {
	if r.done {
		return true, r.err
	}
	if r.total+len(p) > MaxResponse {
		return true, r.fail(fmt.Errorf("%w: больше %d байт", ErrMalformedResponse, MaxResponse))
	}
	r.total += len(p)
	start := len(r.buf)
	r.buf = append(r.buf, p...)
	if err := r.scan(start); err != nil {
		return true, r.fail(err)
	}
	fields, tail := r.seps+1, r.buf[r.tailPos:]
	switch {
	case fields > r.want+1:
		return true, r.fail(fmt.Errorf("%w: больше %d значений", ErrMalformedResponse, r.want))
	case fields == r.want+1 && isSentinelValue(string(tail)):
		r.values = append(r.values, Value(tail))
		r.done = true
		return true, nil
	}
	r.buf = r.buf[:copy(r.buf, tail)]
	r.tailPos = 0
	return false, nil
}

func (r *ResponseReader) scan(start int) error {
	for i := start; i < len(r.buf); i++ {
		c := r.buf[i]
		if c != FieldSep[0] {
			if isControl(rune(c)) {
				return fmt.Errorf("%w: управляющий байт %#02x", ErrMalformedResponse, c)
			}
			continue
		}
		if r.seps == 0 {
			if err := r.checkID(string(r.buf[:i])); err != nil {
				return err
			}
		} else {
			r.values = append(r.values, Value(r.buf[r.tailPos:i]))
		}
		r.seps++
		r.tailPos = i + 1
	}
	return nil
}

// Stalled завершает ответ по дедлайну чтения.
// prt_err последним полем даёт ErrPacketRejected, обрыв ErrTruncatedResponse.
func (r *ResponseReader) Stalled() error {
	if r.done {
		return r.err
	}
	tail := r.buf[r.tailPos:]
	if r.seps >= 1 && string(tail) == PrtErr {
		return r.fail(ErrPacketRejected)
	}
	if r.seps == r.want {
		if isPartialSentinel(string(tail)) {
			return r.fail(fmt.Errorf("%w: ответ сторожевого запроса %q не дочитан", ErrTruncatedResponse, tail))
		}
		r.values = append(r.values, Value(tail))
		r.done = true
		return nil
	}
	return r.fail(fmt.Errorf("%w: %d из %d значений", ErrTruncatedResponse, r.seps, r.want))
}

func (r *ResponseReader) checkID(head string) error {
	id, err := strconv.Atoi(head)
	if err != nil {
		return fmt.Errorf("%w: неверный id %q", ErrMalformedResponse, head)
	}
	if id != r.id {
		return fmt.Errorf("%w: id %d, ожидался %d", ErrMalformedResponse, id, r.id)
	}
	return nil
}

func (r *ResponseReader) fail(err error) error {
	r.done, r.err = true, err
	return err
}

// Response возвращает id и значения со сторожевым, а до конца ответа ошибку.
// Срез значений принадлежит вызывающему.
func (r *ResponseReader) Response() (int, []Value, error) {
	if err := r.ready(); err != nil {
		return 0, nil, err
	}
	return r.id, slices.Clone(r.values), nil
}
