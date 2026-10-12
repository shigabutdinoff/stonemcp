package mindconnect

import (
	"fmt"
	"math"
	"strconv"
)

const (
	intKind = FieldSep + "i" + FieldSep
	strKind = FieldSep + "s" + FieldSep
)

// Arg хранит целый или строковый аргумент вызова функции.
type Arg struct {
	isInt bool
	i     int
	s     string
}

// Int строит целый аргумент; вне диапазона int32 его отвергнет Validate.
func Int(v int) Arg { return Arg{isInt: true, i: v} }

// Str строит строковый аргумент.
func Str(v string) Arg { return Arg{s: v} }

// String возвращает целый аргумент в десятичной записи, строковый как есть.
func (a Arg) String() string {
	if a.isInt {
		return strconv.Itoa(a.i)
	}
	return a.s
}

func (a Arg) validate(owner string) error {
	if a.isInt && (a.i < math.MinInt32 || a.i > math.MaxInt32) {
		return fmt.Errorf("%w: аргумент %d у %q вне диапазона int32", ErrInvalidRequest, a.i, owner)
	}
	if !a.isInt && badText(a.s) {
		return fmt.Errorf("%w: управляющий символ или битый UTF-8 в аргументе %q", ErrInvalidRequest, owner)
	}
	return nil
}

func (a Arg) appendTo(dst []byte) []byte {
	if a.isInt {
		dst = append(dst, intKind...)
		return strconv.AppendInt(dst, int64(a.i), 10)
	}
	dst = append(dst, strKind...)
	return append(dst, a.s...)
}

func (a Arg) size() int {
	if a.isInt {
		return len(intKind) + intLen(a.i)
	}
	return len(strKind) + len(a.s)
}
