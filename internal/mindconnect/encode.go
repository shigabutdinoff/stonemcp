package mindconnect

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrTooLarge означает, что пакет или запрос не укладывается в лимит размера.
var ErrTooLarge = errors.New("mindconnect: пакет слишком большой")

var maxIDDigits = intLen(MaxID)

var sentinelRequest = Var(Sentinel)

var sentinelSize = sentinelRequest.encodedSize()

// EncodeBatch собирает пакет, дописав сторожевой запрос, и читатель ответа.
// maxBytes, равный нулю или больше SafePacket, заменяется на SafePacket.
func EncodeBatch(id int, reqs []Request, maxBytes int) ([]byte, *ResponseReader, error) {
	if id < 1 || id > MaxID {
		return nil, nil, fmt.Errorf("%w: id %d вне диапазона 1..%d", ErrInvalidRequest, id, MaxID)
	}
	if len(reqs) == 0 {
		return nil, nil, fmt.Errorf("%w: пустой набор запросов", ErrInvalidRequest)
	}
	count := len(reqs) + 1
	size := overhead(intLen(id), len(reqs))
	for _, r := range reqs {
		if err := r.Validate(); err != nil {
			return nil, nil, err
		}
		size += r.encodedSize()
	}
	if limit := packetLimit(maxBytes); size > limit {
		return nil, nil, fmt.Errorf("%w: %d байт, лимит %d", ErrTooLarge, size, limit)
	}
	packet := make([]byte, 0, size)
	packet = strconv.AppendInt(packet, int64(id), 10)
	packet = append(packet, FieldSep...)
	packet = strconv.AppendInt(packet, int64(count), 10)
	for _, r := range reqs {
		packet = r.appendFields(packet)
	}
	packet = sentinelRequest.appendFields(packet)
	return packet, NewResponseReader(id, count), nil
}

// Fits проверяет запрос и его место в пакете из одного запроса при любом id.
// Негодный или не влезающий в maxBytes запрос даёт ошибку.
func Fits(r Request, maxBytes int) error {
	_, err := fitsIn(r, packetLimit(maxBytes))
	return err
}

func fitsIn(r Request, limit int) (int, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}
	n := r.encodedSize()
	if alone := overhead(maxIDDigits, 1) + n; alone > limit {
		return 0, fmt.Errorf("%w: запросу %s нужно %d байт, лимит %d", ErrTooLarge, r, alone, limit)
	}
	return n, nil
}

// SplitBatches режет reqs на группы, пакет каждой влезает в maxBytes при
// любом id; негодный или не влезающий в одиночку запрос сразу даёт ошибку.
func SplitBatches(reqs []Request, maxBytes int) ([][]Request, error) {
	var out [][]Request
	limit := packetLimit(maxBytes)
	start, size := 0, 0
	for i, r := range reqs {
		n, err := fitsIn(r, limit)
		if err != nil {
			return nil, err
		}
		if i > start && overhead(maxIDDigits, i-start+1)+size+n > limit {
			out = append(out, reqs[start:i:i])
			start, size = i, 0
		}
		size += n
	}
	if start < len(reqs) {
		out = append(out, reqs[start:len(reqs):len(reqs)])
	}
	return out, nil
}

func overhead(idLen, n int) int {
	return idLen + 1 + intLen(n+1) + sentinelSize
}

func packetLimit(maxBytes int) int {
	if maxBytes <= 0 || maxBytes > SafePacket {
		return SafePacket
	}
	return maxBytes
}

// NextID возвращает id, следующий за id; после MaxID и при id < 1 даёт 1.
func NextID(id int) int {
	if id >= MaxID || id < 1 {
		return 1
	}
	return id + 1
}

const cmdHead = FieldSep + "1" + FieldSep

func (r Request) appendFields(dst []byte) []byte {
	if r.command {
		dst = append(dst, cmdHead...)
		dst = append(dst, r.name...)
		dst = append(dst, FieldSep...)
		return append(dst, r.args[0].s...)
	}
	dst = append(dst, FieldSep...)
	dst = strconv.AppendInt(dst, int64(len(r.args)), 10)
	dst = append(dst, FieldSep...)
	dst = append(dst, r.name...)
	for _, a := range r.args {
		dst = a.appendTo(dst)
	}
	return dst
}

func (r Request) encodedSize() int {
	if r.command {
		return len(cmdHead) + len(r.name) + len(FieldSep) + len(r.args[0].s)
	}
	n := len(FieldSep) + intLen(len(r.args)) + len(FieldSep) + len(r.name)
	for _, a := range r.args {
		n += a.size()
	}
	return n
}

func intLen(v int) int {
	n, u := 1, uint64(v)
	if v < 0 {
		n, u = 2, -uint64(v)
	}
	for u >= 10 {
		u /= 10
		n++
	}
	return n
}
