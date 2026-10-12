package mindconnect

import (
	"fmt"
	"strconv"
	"strings"
)

// Служебные значения ответа игры.
const (
	// CmdDone означает, что игра приняла команду.
	CmdDone = "cmd_done"
	// CmdErr означает, что игра отклонила команду.
	CmdErr = "cmd_err"
	// PrtErr означает, что игра не разобрала пакет; см. ErrPacketRejected.
	PrtErr = "prt_err"
)

const (
	sentinelTrue  = "True"
	sentinelFalse = "False"
	funcErrPrefix = "fnc_err"
)

// Value хранит одно поле ответа.
type Value string

// String возвращает значение как есть.
func (v Value) String() string { return string(v) }

// Int разбирает значение как целое.
func (v Value) Int() (int, error) { return strconv.Atoi(string(v)) }

// Bool разбирает значение как True или False; регистр важен.
func (v Value) Bool() (bool, error) {
	switch v {
	case sentinelTrue:
		return true, nil
	case sentinelFalse:
		return false, nil
	}
	return false, fmt.Errorf("mindconnect: значение %q не True и не False", string(v))
}

// IsError сообщает, код ли это ошибки cmd_err, prt_err или fnc_err с цифрой.
func (v Value) IsError() bool {
	return v == CmdErr || v == PrtErr || isFuncErr(string(v))
}

func isFuncErr(v string) bool {
	rest, ok := strings.CutPrefix(v, funcErrPrefix)
	return ok && len(rest) == 1 && rest[0] >= '0' && rest[0] <= '9'
}

func isSentinelValue(v string) bool {
	return v == sentinelTrue || v == sentinelFalse || isFuncErr(v)
}

func isPartialSentinel(v string) bool {
	return strings.HasPrefix(sentinelTrue, v) || strings.HasPrefix(sentinelFalse, v) || strings.HasPrefix(funcErrPrefix, v)
}
