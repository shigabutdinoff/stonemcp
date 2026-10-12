package engine

import "errors"

// ErrDisconnected означает отсутствие соединения с игрой.
var ErrDisconnected = errors.New("игра не подключена")

// ErrNotTicking означает, что игра не дала результат за отведённое время.
var ErrNotTicking = errors.New("игра не тикает")

// ErrUnsupportedMode означает режим игры, отличный от sync.
var ErrUnsupportedMode = errors.New("режим не поддержан в stonemcp 27.0")
