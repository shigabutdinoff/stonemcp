package mindconnect

import "fmt"

// AppendValues дописывает к dst значения ответа без сторожевого.
// До конца ответа и после ошибки возвращает dst без изменений и ошибку.
func (r *ResponseReader) AppendValues(dst []Value) ([]Value, error) {
	if err := r.ready(); err != nil {
		return dst, err
	}
	return append(dst, r.values[:len(r.values)-1]...), nil
}

func (r *ResponseReader) ready() error {
	if r.err != nil {
		return r.err
	}
	if !r.done {
		return fmt.Errorf("%w: %d из %d значений", ErrIncompleteResponse, r.seps, r.want)
	}
	return nil
}
