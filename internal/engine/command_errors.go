package engine

import (
	"errors"
	"fmt"
)

// ErrQueueFull означает, что очередь команд заполнена.
var ErrQueueFull = errors.New("очередь команд переполнена")

// ErrHoldsFull означает, что число удержаний достигло предела.
var ErrHoldsFull = errors.New("удержаний слишком много")

// ErrExpired означает истечение срока отправки команды.
var ErrExpired = errors.New("команда устарела")

// ErrLocationChanged означает смену локации до отправки команды.
var ErrLocationChanged = errors.New("локация сменилась")

// GameError содержит отказ игры при обработке команды.
type GameError struct {
	// Code содержит код ошибки игры.
	Code string
	// Command содержит полную строку отвергнутой команды.
	Command string
}

// Error возвращает код игры и отвергнутую команду.
func (e *GameError) Error() string {
	return fmt.Sprintf("игра отвергла команду %q с кодом %q", e.Command, e.Code)
}
