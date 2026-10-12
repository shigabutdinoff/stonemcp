package linktest

import (
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	"github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

const (
	readWait  = 2 * time.Second
	quietWait = 50 * time.Millisecond
)

func readSome(t *testing.T, conn net.Conn, buf []byte) []byte {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(readWait))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	return buf[:n]
}

func signal(t *testing.T, conn net.Conn) mindconnect.Phase {
	t.Helper()
	buf := make([]byte, 64)
	var got []byte
	for {
		got = append(got, readSome(t, conn, buf)...)
		phase, version, err := mindconnect.ParseSignal(got)
		switch {
		case errors.Is(err, mindconnect.ErrIncompleteSignal):
			continue
		case err != nil:
			t.Fatalf("сигнал %q: %v", got, err)
		case version != mindconnect.Version:
			t.Fatalf("версия сигнала %q, ожидалось %q", version, mindconnect.Version)
		}
		return phase
	}
}

func ask(t *testing.T, conn net.Conn, id int, reqs []mindconnect.Request) []mindconnect.Value {
	t.Helper()
	values, _ := askCounted(t, conn, id, reqs)
	return values
}

func send(t *testing.T, conn net.Conn, id int, reqs []mindconnect.Request) *mindconnect.ResponseReader {
	t.Helper()
	packet, reader, err := mindconnect.EncodeBatch(id, reqs, mindconnect.SafePacket)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write(packet); err != nil {
		t.Fatal(err)
	}
	return reader
}

func askCounted(t *testing.T, conn net.Conn, id int, reqs []mindconnect.Request) ([]mindconnect.Value, int) {
	t.Helper()
	reader := send(t, conn, id, reqs)
	buf := make([]byte, mindconnect.MaxPacket)
	feeds := 0
	for {
		feeds++
		done, err := reader.Feed(readSome(t, conn, buf))
		if err != nil {
			t.Fatal(err)
		}
		if done {
			break
		}
	}
	gotID, values, err := reader.Response()
	if err != nil {
		t.Fatal(err)
	}
	if gotID != id {
		t.Fatalf("id ответа = %d, ожидалось %d", gotID, id)
	}
	return values, feeds
}

func askStalled(t *testing.T, conn net.Conn, id int, reqs []mindconnect.Request) *mindconnect.ResponseReader {
	t.Helper()
	return awaitStalled(t, conn, send(t, conn, id, reqs))
}

func awaitStalled(t *testing.T, conn net.Conn, reader *mindconnect.ResponseReader) *mindconnect.ResponseReader {
	t.Helper()
	buf := make([]byte, mindconnect.MaxPacket)
	wait := readWait
	for {
		conn.SetReadDeadline(time.Now().Add(wait))
		n, err := conn.Read(buf)
		if n > 0 {
			if done, ferr := reader.Feed(buf[:n]); done {
				t.Fatalf("Feed: завершён с err=%v до того, как игра замолчала, ожидалось, что решение останется за Stalled", ferr)
			}
			wait = quietWait
		}
		switch {
		case err == nil:
		case errors.Is(err, os.ErrDeadlineExceeded) && wait == quietWait:
			reader.Stalled()
			return reader
		default:
			t.Fatal(err)
		}
	}
}

func wantViolation(t *testing.T, g *fakegame.Game, about string) {
	t.Helper()
	if v := g.Violations(); len(v) != 1 || !strings.Contains(v[0], about) {
		t.Errorf("violations = %v, ожидалось одно нарушение про %q", v, about)
	}
}

func endStep(t *testing.T, conn net.Conn) {
	t.Helper()
	if _, err := conn.Write([]byte(mindconnect.EndStep)); err != nil {
		t.Fatal(err)
	}
}
