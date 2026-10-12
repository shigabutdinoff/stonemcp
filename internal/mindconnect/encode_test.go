package mindconnect

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestEncodeBatch(t *testing.T) {
	tests := []struct {
		name string
		id   int
		reqs []Request
		want string
	}{
		{"variable", 1, []Request{Var("hp")}, "1\x1f2\x1f0\x1fhp\x1f0\x1fai.enabled"},
		{"spike batch", 1, []Request{Var("time"), Var("hp"), Var("maxhp"), Var("loc.id"), Var("foe"), Var("foe.distance"), Var("pos.x")},
			"1\x1f8\x1f0\x1ftime\x1f0\x1fhp\x1f0\x1fmaxhp\x1f0\x1floc.id\x1f0\x1ffoe\x1f0\x1ffoe.distance\x1f0\x1fpos.x\x1f0\x1fai.enabled"},
		{"function with int and string args", 42, []Request{Var("draw.GetSymbol", Int(3), Int(7)), Var("item.GetCooldown", Str("bardiche"))},
			"42\x1f3\x1f2\x1fdraw.GetSymbol\x1fi\x1f3\x1fi\x1f7\x1f1\x1fitem.GetCooldown\x1fs\x1fbardiche\x1f0\x1fai.enabled"},
		{"command", 2, []Request{Cmd(">", "`0,1,#green,hi")}, "2\x1f2\x1f1\x1f>\x1f`0,1,#green,hi\x1f0\x1fai.enabled"},
		{"command and variable", 9999, []Request{Cmd("equipL", "sword"), Var("hp")}, "9999\x1f3\x1f1\x1fequipL\x1fsword\x1f0\x1fhp\x1f0\x1fai.enabled"},
		{"faerie and profile commands", 4, []Request{Cmd("equipF", "wisp"), Cmd("profile", "Main")},
			"4\x1f3\x1f1\x1fequipF\x1fwisp\x1f1\x1fprofile\x1fMain\x1f0\x1fai.enabled"},
		{"at sign in a function argument", 5, []Request{Var("var.set", Str("mail"), Str("a@b"))},
			"5\x1f2\x1f2\x1fvar.set\x1fs\x1fmail\x1fs\x1fa@b\x1f0\x1fai.enabled"},
		{"int32 bounds", 3, []Request{Var("f", Int(math.MaxInt32), Int(math.MinInt32))}, "3\x1f2\x1f2\x1ff\x1fi\x1f2147483647\x1fi\x1f-2147483648\x1f0\x1fai.enabled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reader, err := EncodeBatch(tt.id, tt.reqs, SafePacket)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("получено  %q\nожидалось %q", got, tt.want)
			}
			answer := strconv.Itoa(tt.id) + strings.Repeat(FieldSep+"True", len(tt.reqs)+1)
			if done, err := reader.Feed([]byte(answer)); !done || err != nil {
				t.Errorf("читатель ответа %q: done=%v err=%v", answer, done, err)
			}
		})
	}
}

func TestEncodeBatchRejects(t *testing.T) {
	tests := []struct {
		name string
		id   int
		reqs []Request
		want error
	}{
		{"id zero", 0, []Request{Var("hp")}, ErrInvalidRequest},
		{"id too big", 10000, []Request{Var("hp")}, ErrInvalidRequest},
		{"empty batch", 1, nil, ErrInvalidRequest},
		{"empty name", 1, []Request{Var("")}, ErrInvalidRequest},
		{"separator in name", 1, []Request{Var("h\x1fp")}, ErrInvalidRequest},
		{"separator in arg", 1, []Request{Cmd("equip", "a\x1fb")}, ErrInvalidRequest},
		{"end-of-step in arg", 1, []Request{Cmd(">", "a\x03")}, ErrInvalidRequest},
		{"signal prefix in arg", 1, []Request{Var("f", Str("\x06"))}, ErrInvalidRequest},
		{"newline in arg", 1, []Request{Cmd(">", "a\nb")}, ErrInvalidRequest},
		{"invalid UTF-8 in name", 1, []Request{Var("h\xffp")}, ErrInvalidRequest},
		{"invalid UTF-8 in arg", 1, []Request{Cmd(">", "a\xc3")}, ErrInvalidRequest},
		{"int above int32", 1, []Request{Var("draw.GetSymbol", Int(math.MaxInt32+1), Int(0))}, ErrInvalidRequest},
		{"int below int32", 1, []Request{Var("draw.GetSymbol", Int(0), Int(math.MinInt32-1))}, ErrInvalidRequest},
		{"unknown command", 1, []Request{Cmd("fly", "up")}, ErrInvalidRequest},
		{"command read as variable", 1, []Request{Var("equip")}, ErrInvalidRequest},
		{"expression in a print", 1, []Request{Cmd(">", "hp @hp@")}, ErrInvalidRequest},
		{"at sign in an item name", 1, []Request{Cmd("equipR", "a@b")}, ErrInvalidRequest},
		{"at sign in a brew", 1, []Request{Cmd("brew", "@")}, ErrInvalidRequest},
		{"profile read as variable", 1, []Request{Var("profile")}, ErrInvalidRequest},
		{"equipF called as function", 1, []Request{Var("equipF", Str("wisp"))}, ErrInvalidRequest},
		{"command name in other case", 1, []Request{Cmd("Equip", "sword")}, ErrInvalidRequest},
		{"larger than the safe packet", 1, []Request{Cmd(">", strings.Repeat("x", SafePacket))}, ErrTooLarge},
		{"larger than the game buffer", 1, []Request{Cmd(">", strings.Repeat("x", MaxPacket))}, ErrTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := EncodeBatch(tt.id, tt.reqs, SafePacket); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, ожидалось %v", err, tt.want)
			}
		})
	}
}

func TestEncodedSizeMatchesEncoding(t *testing.T) {
	reqs := []Request{
		Var("hp"),
		Var("draw.GetSymbol", Int(-10), Int(7)),
		Var("item.GetCooldown", Str("bardiche")),
		Var("f", Int(0), Str(""), Int(12345)),
		Var("f", Int(math.MinInt), Int(math.MaxInt)),
		Cmd(">", "`0,1,#green,hi"),
		Cmd("equipL", ""),
	}
	for _, r := range reqs {
		if got, want := r.encodedSize(), len(r.appendFields(nil)); got != want {
			t.Errorf("%s: encodedSize = %d, поля занимают %d", r, got, want)
		}
	}

	packet, _, err := EncodeBatch(1, []Request{Var("draw.GetSymbol", Int(3), Int(7))}, SafePacket)
	if err != nil {
		t.Fatal(err)
	}
	header := len("1") + 1 + len("2")
	if got, want := Var("draw.GetSymbol", Int(3), Int(7)).encodedSize(), len(packet)-header-sentinelSize; got != want {
		t.Errorf("encodedSize = %d, вклад в пакет %d", got, want)
	}
}

func TestNextID(t *testing.T) {
	for in, want := range map[int]int{0: 1, 1: 2, 9998: 9999, 9999: 1} {
		if got := NextID(in); got != want {
			t.Errorf("NextID(%d) = %d, ожидалось %d", in, got, want)
		}
	}
}
