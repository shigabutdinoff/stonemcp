package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/mcpserver"
)

func TestServeWithoutGame(t *testing.T) {
	setVersion(t, "27.0-beta.2", "c4e1d20")
	want := "27.0-beta.2 (c4e1d20)"
	l := start(t, "--addr", noGame)
	if got := l.InitializeResult().ServerInfo.Version; got != want {
		t.Errorf("версия при подключении = %q, ожидалось %q", got, want)
	}
	res, err := l.CallTool(context.Background(), &mcp.CallToolParams{Name: "stone_status"})
	if err != nil {
		t.Fatalf("stone_status: %v", err)
	}
	var st mcpserver.StatusOut
	raw, _ := json.Marshal(res.StructuredContent)
	if err := json.Unmarshal(raw, &st); err != nil {
		t.Fatalf("не разобран ответ %s: %v", raw, err)
	}
	if st.Connected || st.Mode != "sync" || st.ServerVersion != want || st.Hint != mcpserver.Hint {
		t.Errorf("статус = %+v", st)
	}
	l.stdin.Close()
	l.want(t, 0)
	if !strings.Contains(l.stderr.String(), "stonemcp запущен") {
		t.Errorf("stderr = %q, ожидался журнал запуска", l.stderr.String())
	}
}

func TestLogFileAndGameAddr(t *testing.T) {
	t.Parallel()
	ln := listen(t)
	path := filepath.Join(t.TempDir(), "s.log")
	l := start(t, "--addr", ln.Addr().String(), "--log-file", path)
	conn := accept(t, ln)
	l.stdin.Close()
	l.wantReleased(t, conn)
	journal, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(journal), "адрес="+ln.Addr().String()) {
		t.Errorf("журнал = %q, err = %v", journal, err)
	}
	if l.stderr.Len() != 0 {
		t.Errorf("stderr = %q, ожидался пустой", l.stderr.String())
	}
}

func TestReconnectFromClient(t *testing.T) {
	t.Parallel()
	ln := listen(t)
	for range 2 {
		l := start(t, "--addr", ln.Addr().String())
		conn := accept(t, ln)
		l.stdin.Close()
		l.wantReleased(t, conn)
	}
}

func TestLogFileError(t *testing.T) {
	var stderr strings.Builder
	path := filepath.Join(t.TempDir(), "нет", "s.log")
	code := run([]string{"--addr", noGame, "--log-file", path}, noEnv, nil, nil, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "stonemcp: файл журнала") {
		t.Errorf("код = %d, stderr = %q, ожидались 1 и ошибка файла журнала", code, stderr.String())
	}
}

func TestServeErrorToLogFile(t *testing.T) {
	t.Parallel()
	var stderr strings.Builder
	path := filepath.Join(t.TempDir(), "s.log")
	code := run([]string{"--addr", noGame, "--log-file", path}, noEnv, brokenInput(), &buffer{}, &stderr)
	journal, err := os.ReadFile(path)
	if code != 1 || err != nil || !strings.Contains(string(journal), "stonemcp остановлен") {
		t.Errorf("код = %d, журнал = %q, err = %v, ожидались 1 и ошибка в журнале", code, journal, err)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, ожидался пустой", stderr.String())
	}
}

type failingTransport struct{ err error }

func (f failingTransport) Connect(context.Context) (mcp.Connection, error) { return nil, f.err }

func TestPlayWrapsClientError(t *testing.T) {
	cause := errors.New("транспорт недоступен")
	o := options{settings: settings{Mode: "sync", Addr: noGame}}
	err := play(context.Background(), o, failingTransport{cause}, slog.New(slog.DiscardHandler))
	if !errors.Is(err, cause) || err.Error() != cause.Error()+": обмен с клиентом MCP" {
		t.Errorf("err = %v, ожидалась обёрнутая ошибка транспорта", err)
	}
}
