package linktest

import (
	"errors"
	"slices"
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	"github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestRoundTripFrame(t *testing.T) {
	g := fakegame.Start(t)
	g.Set("loc.name", "Sea Cave")
	g.SetScreen([]string{"#-#", "|.|"})
	conn := g.Dial(t)

	if phase := signal(t, conn); phase != mindconnect.PhasePre {
		t.Fatalf("phase = %s, ожидалось pre", phase)
	}
	values := ask(t, conn, 1, []mindconnect.Request{
		mindconnect.Var("hp"),
		mindconnect.Var("loc.name"),
		mindconnect.Var("draw.GetSymbol", mindconnect.Int(1), mindconnect.Int(1)),
		mindconnect.Var("nope"),
		mindconnect.Cmd("equipL", "sword"),
	})
	if len(values) != 6 {
		t.Fatalf("values = %q, ожидалось пять ответов и ответ сторожевого запроса", values)
	}
	if !values[3].IsError() {
		t.Errorf("неизвестная переменная ответила %q, ожидался код ошибки", values[3])
	}
	want := []mindconnect.Value{"80", "Sea Cave", ".", values[3], mindconnect.CmdDone, "True"}
	if !slices.Equal(values, want) {
		t.Errorf("values = %q, ожидалось %q", values, want)
	}
	endStep(t, conn)

	if phase := signal(t, conn); phase != mindconnect.PhasePost {
		t.Fatalf("phase = %s, ожидалось post", phase)
	}
	cmds := g.CommandsNamed("equipL")
	if len(cmds) != 1 || cmds[0].Arg != "sword" || cmds[0].Phase != "pre" {
		t.Errorf("commands = %+v", cmds)
	}
}

func TestRoundTripSplitResponse(t *testing.T) {
	g := fakegame.Start(t)
	g.SplitResponses(true)
	conn := g.Dial(t)

	for frame := range 20 {
		signal(t, conn)
		values, feeds := askCounted(t, conn, 17, []mindconnect.Request{
			mindconnect.Var("hp"),
			mindconnect.Var("loc.id"),
		})
		if want := []mindconnect.Value{"80", "rocky_plateau", "True"}; !slices.Equal(values, want) {
			t.Fatalf("кадр %d: values = %q, ожидалось %q", frame, values, want)
		}
		if feeds >= 2 {
			return
		}
		endStep(t, conn)
	}
	t.Fatal("ответ так и не пришёл частями")
}

func TestRoundTripFragmentedPacket(t *testing.T) {
	g := fakegame.Start(t)
	g.AllowViolations()
	conn := g.Dial(t)
	signal(t, conn)

	packet, reader, err := mindconnect.EncodeBatch(3, []mindconnect.Request{mindconnect.Var("hp")}, mindconnect.SafePacket)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write(packet[:len(packet)/2]); err != nil {
		t.Fatal(err)
	}
	awaitStalled(t, conn, reader)
	if _, _, err := reader.Response(); !errors.Is(err, mindconnect.ErrPacketRejected) {
		t.Fatalf("Response err = %v, ожидалось, что первая половина отвергнута как отдельный пакет", err)
	}
	wantViolation(t, g, "искажённый пакет")
}

func TestRoundTripRejectedPacket(t *testing.T) {
	g := fakegame.Start(t)
	g.RejectPackets(true)
	conn := g.Dial(t)
	signal(t, conn)

	reader := askStalled(t, conn, 4, []mindconnect.Request{mindconnect.Var("hp"), mindconnect.Var("maxhp")})
	if _, _, err := reader.Response(); !errors.Is(err, mindconnect.ErrPacketRejected) {
		t.Fatalf("Response err = %v, ожидалось ErrPacketRejected", err)
	}

	g.RejectPackets(false)
	if values := ask(t, conn, 5, []mindconnect.Request{mindconnect.Var("hp")}); values[0] != "80" {
		t.Errorf("values = %q, ожидалось, что связь переживёт отказ", values)
	}
}

func TestRoundTripAbortedTick(t *testing.T) {
	g := fakegame.Start(t)
	g.AllowViolations()
	conn := g.Dial(t)
	signal(t, conn)

	reader := send(t, conn, 6, []mindconnect.Request{mindconnect.Var("draw.GetSymbol", mindconnect.Int(1))})
	got := readSome(t, conn, make([]byte, mindconnect.MaxPacket))
	if done, err := reader.Feed(got); !done || !errors.Is(err, mindconnect.ErrMalformedResponse) {
		t.Fatalf("Feed(%q): done=%v err=%v, ожидалось, что следующий сигнал пойман как управляющий байт", got, done, err)
	}
	wantViolation(t, g, "исключение в игре")
}

func TestRoundTripUnusualSentinel(t *testing.T) {
	g := fakegame.Start(t)
	g.Set("ai.enabled", "maybe")
	conn := g.Dial(t)
	signal(t, conn)

	reader := askStalled(t, conn, 8, []mindconnect.Request{mindconnect.Var("hp"), mindconnect.Var("loc.id")})
	gotID, values, err := reader.Response()
	if err != nil {
		t.Fatalf("Response err = %v, ожидалось, что ответ принят целиком", err)
	}
	want := []mindconnect.Value{"80", "rocky_plateau", "maybe"}
	if gotID != 8 || !slices.Equal(values, want) {
		t.Errorf("id = %d values = %q, ожидалось 8 и %q", gotID, values, want)
	}
}
