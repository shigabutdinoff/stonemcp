//go:build !release

package fakegame

import (
	"fmt"
	"io"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const stopWait = 5 * time.Second

// Game управляет одним процессом тестовой игры до конца теста.
type Game struct {
	tb    testing.TB
	cmd   *exec.Cmd
	rpc   *rpc.Client
	addr  string
	allow atomic.Bool
	once  sync.Once
}

type pipe struct {
	io.Reader
	io.WriteCloser
}

// Start запускает игру до конца теста; нарушение валит его без AllowViolations.
func Start(tb testing.TB) *Game {
	tb.Helper()
	path, err := binaryPath()
	if err != nil {
		tb.Fatal(err)
	}
	g, err := launch(tb, path)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(g.stop)
	return g
}

func launch(tb testing.TB, path string) (*Game, error) {
	cmd := exec.Command(path)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%w: запуск тестовой игры для %s/%s", err, runtime.GOOS, runtime.GOARCH)
	}
	g := &Game{tb: tb, cmd: cmd, rpc: rpc.NewClient(pipe{stdout, stdin})}
	if err := g.rpc.Call("Game.Addr", struct{}{}, &g.addr); err != nil {
		g.quit()
		return nil, fmt.Errorf("%w: тестовая игра не ответила после запуска", err)
	}
	return g, nil
}

func (g *Game) stop() {
	g.once.Do(func() {
		var r report
		g.call("Report", struct{}{}, &r)
		if len(r.Violations) > 0 && !g.allow.Load() {
			g.tb.Errorf("клиент нарушил протокол: %v", r.Violations)
		}
		for _, failure := range r.Failures {
			g.tb.Error(failure)
		}
		g.quit()
	})
}

func (g *Game) quit() {
	_ = g.rpc.Close()
	kill := time.AfterFunc(stopWait, func() { _ = g.cmd.Process.Kill() })
	defer kill.Stop()
	_ = g.cmd.Wait()
}
