package mindconnect

import (
	"errors"
	"testing"
)

var errOther = errors.New("посторонняя ошибка")

func TestParseSignal(t *testing.T) {
	tests := []struct {
		in      string
		phase   Phase
		version string
		err     error
	}{
		{in: "\x06pre0.3", phase: PhasePre, version: "0.3"},
		{in: "\x06post0.3", phase: PhasePost, version: "0.3"},
		{in: "\x06pre1.12", phase: PhasePre, version: "1.12"},
		{in: "\x060.3", err: ErrUnsupportedSignal},
		{in: "\x06sync0.3", err: ErrUnsupportedSignal},
		{in: "\x06pre", err: ErrIncompleteSignal},
		{in: "\x06post", err: ErrIncompleteSignal},
		{in: "\x06pre0", err: ErrIncompleteSignal},
		{in: "\x06pre0.", err: ErrIncompleteSignal},
		{in: "\x06p", err: ErrIncompleteSignal},
		{in: "\x06pos", err: ErrIncompleteSignal},
		{in: "\x06", err: ErrIncompleteSignal},
		{in: "", err: ErrIncompleteSignal},
		{in: "\x06pre0.x", err: errOther},
		{in: "\x06pre..3", err: errOther},
		{in: "pre0.3", err: errOther},
		{in: "1\x1fcmd_done", err: errOther},
	}
	for _, tt := range tests {
		phase, version, err := ParseSignal([]byte(tt.in))
		if (err != nil) != (tt.err != nil) {
			t.Errorf("%q: err = %v, ожидалось %v", tt.in, err, tt.err)
			continue
		}
		for _, sentinel := range []error{ErrUnsupportedSignal, ErrIncompleteSignal} {
			if got, want := errors.Is(err, sentinel), tt.err == sentinel; got != want {
				t.Errorf("%q: errors.Is(err, %v) = %v, ожидалось %v", tt.in, sentinel, got, want)
			}
		}
		if phase != tt.phase || version != tt.version {
			t.Errorf("%q: получено (%v, %q), ожидалось (%v, %q)", tt.in, phase, version, tt.phase, tt.version)
		}
	}
}

func TestParseSignalByteByByte(t *testing.T) {
	for _, full := range []string{"\x06pre" + Version, "\x06post" + Version} {
		for n := 1; n < len(full); n++ {
			_, _, err := ParseSignal([]byte(full[:n]))
			if !errors.Is(err, ErrIncompleteSignal) {
				t.Errorf("%q: err = %v, ожидалось ErrIncompleteSignal", full[:n], err)
			}
		}
		phase, version, err := ParseSignal([]byte(full))
		if err != nil || phase == 0 || version != Version {
			t.Errorf("%q: получено (%v, %q), err = %v", full, phase, version, err)
		}
	}
}

func TestParseStalledSignal(t *testing.T) {
	tests := []struct {
		in          string
		phase       Phase
		version     string
		wantErr     bool
		unsupported bool
	}{
		{in: "\x06pre0.3", phase: PhasePre, version: "0.3"},
		{in: "\x06post0.3", phase: PhasePost, version: "0.3"},
		{in: "\x06pre1", phase: PhasePre, version: "1"},
		{in: "\x06post0", phase: PhasePost, version: "0"},
		{in: "\x06pre", wantErr: true},
		{in: "\x06pre0.", wantErr: true},
		{in: "\x06p", wantErr: true},
		{in: "\x06", wantErr: true},
		{in: "", wantErr: true},
		{in: "\x06pre0.x", wantErr: true},
		{in: "\x060.3", wantErr: true, unsupported: true},
		{in: "pre0.3", wantErr: true},
	}
	for _, tt := range tests {
		phase, version, err := ParseStalledSignal([]byte(tt.in))
		if (err != nil) != tt.wantErr {
			t.Errorf("%q: err = %v, ожидалось wantErr = %v", tt.in, err, tt.wantErr)
			continue
		}
		if errors.Is(err, ErrIncompleteSignal) {
			t.Errorf("%q: err = %v, после дедлайна чтения ErrIncompleteSignal не ожидался", tt.in, err)
		}
		if got := errors.Is(err, ErrUnsupportedSignal); got != tt.unsupported {
			t.Errorf("%q: errors.Is(err, ErrUnsupportedSignal) = %v, ожидалось %v", tt.in, got, tt.unsupported)
		}
		if phase != tt.phase || version != tt.version {
			t.Errorf("%q: получено (%v, %q), ожидалось (%v, %q)", tt.in, phase, version, tt.phase, tt.version)
		}
	}
}
