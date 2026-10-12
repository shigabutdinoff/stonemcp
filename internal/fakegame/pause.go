//go:build !release

package fakegame

import "context"

// Pause останавливает кадры на границе кадра до Resume; соединение остаётся.
func (g *Game) Pause() { g.do("Pause", struct{}{}) }

// Resume возобновляет кадры после Pause и освобождает ожидание паузы.
func (g *Game) Resume() { g.do("Resume", struct{}{}) }

// PauseAndWait запрашивает паузу и ждёт завершения текущего кадра.
// Достигнутая пауза возвращает nil, иначе отмена возвращает ошибку ctx.
// Resume до конца кадра освобождает ожидание.
func (g *Game) PauseAndWait(ctx context.Context) error {
	call := g.rpc.Go("Game.PauseAndWait", struct{}{}, &struct{}{}, nil)
	select {
	case <-call.Done:
		return call.Error
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PauseAfterCommand подменяет ai.enabled и ставит паузу, как только игра
// получила команду name. Канал закрывается после запроса паузы.
func (g *Game) PauseAfterCommand(name string) <-chan struct{} {
	g.do("PauseAfterCommand", name)
	paused := make(chan struct{})
	call := g.rpc.Go("Game.AwaitPause", struct{}{}, &struct{}{}, nil)
	go func() {
		if done := <-call.Done; done.Error == nil {
			close(paused)
		}
	}()
	return paused
}
