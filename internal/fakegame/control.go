//go:build !release

package fakegame

import (
	"net"
	"testing"
)

// Addr возвращает адрес, на котором слушает игра.
func (g *Game) Addr() string { return g.addr }

// Dial подключает клиента к игре и закрывает соединение в конце теста.
func (g *Game) Dial(tb testing.TB) net.Conn {
	tb.Helper()
	conn, err := net.Dial("tcp", g.addr)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { _ = conn.Close() })
	return conn
}

// Close останавливает игру, сохраняя журнал; повторный вызов безопасен.
func (g *Game) Close() { g.do("Close", struct{}{}) }

// DropConn рвёт соединение, как перезапуск игры; без клиента валит тест.
func (g *Game) DropConn() { g.do("DropConn", struct{}{}) }

// SetMode задаёт sync, exclusive или async со следующего игрового кадра.
// Неизвестное имя проваливает тест.
func (g *Game) SetMode(mode string) { g.do("SetMode", mode) }

// SetVersion меняет версию протокола в сигналах кадра.
func (g *Game) SetVersion(v string) { g.do("SetVersion", v) }

// Reject включает и выключает ответ cmd_err на команду name.
func (g *Game) Reject(name string, on bool) { g.do("Reject", toggle{name, on}) }

// RejectPackets отвечает на каждый пакет id и prt_err и держит соединение.
func (g *Game) RejectPackets(on bool) { g.do("RejectPackets", on) }

// SplitResponses заставляет игру отдавать каждый ответ двумя записями.
func (g *Game) SplitResponses(on bool) { g.do("SplitResponses", on) }

// Frame возвращает номер текущего кадра.
func (g *Game) Frame() int {
	var frame int
	g.call("Frame", struct{}{}, &frame)
	return frame
}
