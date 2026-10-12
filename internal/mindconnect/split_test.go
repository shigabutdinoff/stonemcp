package mindconnect

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestSplitBatches(t *testing.T) {
	reqs := screenRequests()
	groups, err := SplitBatches(reqs, SafePacket)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) < 2 {
		t.Fatalf("получено групп: %d, ожидалось несколько", len(groups))
	}
	total := 0
	for i, g := range groups {
		packet, _, err := EncodeBatch(1, g, SafePacket)
		if err != nil {
			t.Fatalf("группа %d: %v", i, err)
		}
		if len(packet) > SafePacket {
			t.Errorf("группа %d занимает %d байт, лимит %d", i, len(packet), SafePacket)
		}
		for j, r := range g {
			want := fmt.Sprintf("draw.GetSymbol(%d, %d)", total%screenW, total/screenW)
			if got := r.String(); got != want {
				t.Fatalf("порядок нарушен: в группе %d запрос %d равен %s, ожидалось %s", i, j, got, want)
			}
			total++
		}
	}
	if total != len(reqs) {
		t.Errorf("в группах %d запросов, ожидалось %d", total, len(reqs))
	}
}

func TestSplitBatchesEdgeCases(t *testing.T) {
	if got, err := SplitBatches(nil, SafePacket); got != nil || err != nil {
		t.Errorf("вход nil: получено %v, err = %v", got, err)
	}
	big := Cmd(">", strings.Repeat("x", 100))
	groups, err := SplitBatches([]Request{Var("hp"), big, Var("maxhp")}, 50)
	if groups != nil || !errors.Is(err, ErrTooLarge) {
		t.Fatalf("groups = %v, err = %v, ожидалось ErrTooLarge для запроса, который никогда не поместится", groups, err)
	}
	if _, _, err := EncodeBatch(1, []Request{big}, 50); !errors.Is(err, ErrTooLarge) {
		t.Errorf("EncodeBatch err = %v, ожидалось то же решение, что у SplitBatches", err)
	}
}

func TestSplitBatchesKeepsWhatFits(t *testing.T) {
	const limit = 50
	r := Cmd(">", strings.Repeat("x", 23))
	groups, err := SplitBatches([]Request{r}, limit)
	if err != nil || len(groups) != 1 {
		t.Fatalf("groups = %v, err = %v, ожидалась одна группа для запроса, который помещается", groups, err)
	}
	packet, _, err := EncodeBatch(MaxID, groups[0], limit)
	if err != nil {
		t.Fatalf("EncodeBatch = %v, ожидалось, что группа закодирована", err)
	}
	if len(packet) > limit {
		t.Errorf("пакет занимает %d байт, лимит %d", len(packet), limit)
	}
}

func TestSplitBatchesRejectsInvalidRequest(t *testing.T) {
	for _, tt := range []struct {
		name string
		bad  Request
	}{
		{"command without an argument", Request{name: ">", command: true}},
		{"unknown command", Cmd("fly", "up")},
		{"command read as a variable", Var("equip")},
		{"separator in an argument", Var("f", Str("a\x1fb"))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reqs := []Request{Var("hp"), tt.bad, Var("maxhp")}
			groups, err := SplitBatches(reqs, SafePacket)
			if groups != nil || !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("groups = %v, err = %v, ожидалось решение до отправки первого пакета", groups, err)
			}
			if _, _, err := EncodeBatch(1, reqs, SafePacket); !errors.Is(err, ErrInvalidRequest) {
				t.Errorf("EncodeBatch err = %v, ожидалось то же решение, что у SplitBatches", err)
			}
		})
	}
}

func TestSplitBatchesAndEncodeShareTheLimit(t *testing.T) {
	reqs := screenRequests()[:300]
	for _, limit := range []int{64, 200, 1000, SafePacket, MaxPacket} {
		groups, err := SplitBatches(reqs, limit)
		if err != nil {
			t.Fatalf("лимит %d: %v", limit, err)
		}
		for i, g := range groups {
			packet, _, err := EncodeBatch(1, g, limit)
			if err != nil {
				t.Fatalf("лимит %d, группа %d: %v", limit, i, err)
			}
			if want := packetLimit(limit); len(packet) > want {
				t.Errorf("лимит %d, группа %d: пакет занимает %d байт", limit, i, len(packet))
			}
		}
	}
}

func TestFits(t *testing.T) {
	if err := Fits(Var("hp"), SafePacket); err != nil {
		t.Fatalf("обычный запрос не принят, err = %v", err)
	}
	if err := Fits(Var(""), SafePacket); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("ошибка пустого имени = %v, ожидалось %v", err, ErrInvalidRequest)
	}
	huge := Cmd(">", strings.Repeat("x", SafePacket))
	if err := Fits(huge, SafePacket); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("ошибка большого запроса = %v, ожидалось %v", err, ErrTooLarge)
	}
	if err := Fits(Var("hp"), 8); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("ошибка малого лимита = %v, ожидалось %v", err, ErrTooLarge)
	}
}
