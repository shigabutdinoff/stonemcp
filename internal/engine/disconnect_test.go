package engine

import (
	"errors"
	"fmt"
	"testing"
)

func TestDisconnectPreservesUnsupportedModeUntilConnected(t *testing.T) {
	e := New(Config{})
	unavailable := errors.New("подключение отклонено")
	for _, cause := range []error{ErrUnsupportedMode, unavailable} {
		if got := e.disconnect(cause); !errors.Is(got, ErrUnsupportedMode) {
			t.Fatalf("причина неподдержанного режима потеряна: %v", got)
		}
		if _, err := e.Snapshot(); !errors.Is(err, ErrUnsupportedMode) {
			t.Fatalf("ошибка состояния после разрыва: %v", err)
		}
		if status := e.Status(); status.Connected || status.LastError != ErrUnsupportedMode.Error() {
			t.Fatalf("статус после разрыва: %+v", status)
		}
	}
	e.connected("0.3")
	if status := e.Status(); !status.Connected || status.LastError != "" {
		t.Fatalf("статус после восстановления: %+v", status)
	}
	if got := e.disconnect(unavailable); !errors.Is(got, unavailable) {
		t.Fatalf("новая причина разрыва потеряна: %v", got)
	}
	if _, err := e.Snapshot(); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("ошибка нового разрыва: %v", err)
	}
}

func TestDisconnectDoesNotClassifyErrorsByText(t *testing.T) {
	e := New(Config{})
	cause := errors.New(ErrUnsupportedMode.Error())
	if got := e.disconnect(cause); got != cause || errors.Is(got, ErrUnsupportedMode) {
		t.Fatalf("причина разрыва подменена по тексту: %v", got)
	}
}

func TestDisconnectRecognizesWrappedUnsupportedMode(t *testing.T) {
	e := New(Config{})
	wrapped := fmt.Errorf("сигнал игры: %w", ErrUnsupportedMode)
	for _, cause := range []error{wrapped, errors.New("подключение отклонено")} {
		if got := e.disconnect(cause); got != wrapped {
			t.Fatalf("обёрнутая причина неподдержанного режима потеряна: %v", got)
		}
		if _, err := e.Snapshot(); !errors.Is(err, ErrUnsupportedMode) {
			t.Fatalf("ошибка состояния после обёрнутой причины: %v", err)
		}
	}
}
