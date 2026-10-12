package engine

import (
	"log/slog"
	"net"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

var pauseAttr = regexp.MustCompile(`пауза=(\S+)`)

type journal struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (j *journal) Write(p []byte) (int, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.buf.Write(p)
}

func (j *journal) pauses() []string {
	j.mu.Lock()
	defer j.mu.Unlock()
	var got []string
	for _, m := range pauseAttr.FindAllStringSubmatch(j.buf.String(), -1) {
		got = append(got, m[1])
	}
	return got
}

func journaled(cfg Config) (Config, *journal) {
	j := &journal{}
	cfg.Logger = slog.New(slog.NewTextHandler(j, nil))
	cfg.ReconnectDelays = []time.Duration{5 * time.Millisecond, 10 * time.Millisecond}
	return cfg, j
}

func TestAbsentGamePausesGrow(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
	cfg, j := journaled(Config{Addr: ln.Addr().String()})
	_, stop := launchEngine(t, cfg)
	eventually(t, "три паузы", func() bool { return len(j.pauses()) >= 3 })
	stop()
	if got := j.pauses()[:3]; !slices.Equal(got, []string{"5ms", "10ms", "10ms"}) {
		t.Fatalf("паузы = %v, ожидались 5ms, 10ms, 10ms", got)
	}
}

func TestConnectedSessionResetsPause(t *testing.T) {
	g := fakegame.Start(t)
	cfg, j := journaled(Config{})
	e, stop := startStoppable(t, g, cfg)
	for i := range 2 {
		g.DropConn()
		eventually(t, "обрыв", func() bool { return len(j.pauses()) > i })
		eventually(t, "переподключение", func() bool { _, err := e.Snapshot(); return err == nil })
	}
	stop()
	if got := j.pauses(); !slices.Equal(got, []string{"5ms", "5ms"}) {
		t.Fatalf("паузы = %v, ожидались две по 5ms", got)
	}
}

func TestUnsupportedModeTakesLastPause(t *testing.T) {
	g := fakegame.Start(t)
	cfg, j := journaled(Config{})
	e, stop := startStoppable(t, g, cfg)
	g.SetMode("async")
	eventually(t, "неподдержанный режим", func() bool { return len(j.pauses()) > 0 })
	stop()
	if got := j.pauses()[0]; got != "10ms" || e.Status().LastError != ErrUnsupportedMode.Error() {
		t.Fatalf("пауза = %s, статус = %+v, ожидалась 10ms", got, e.Status())
	}
}

func TestReconnectDelaysConfig(t *testing.T) {
	own := []time.Duration{time.Second}
	c := Config{ReconnectDelays: own}.defaults()
	own[0] = 0
	if c.ReconnectDelays[0] != time.Second {
		t.Errorf("паузы = %v, ожидалась копия слайса вызывающего", c.ReconnectDelays)
	}
	want := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}
	if got := (Config{}).defaults().ReconnectDelays; !slices.Equal(got, want) {
		t.Errorf("паузы по умолчанию = %v, ожидались %v", got, want)
	}
	if err := (Config{ReconnectDelays: []time.Duration{}}).defaults().validate(); err == nil {
		t.Error("приняты пустые паузы")
	}
}
