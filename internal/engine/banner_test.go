package engine

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

const bannerText = "stonemcp 27.0 (27A41)"

func wideScreen(n int) []string { return []string{strings.Repeat(".", n)} }

func bannerAt(x string) string { return "`" + x + ",0," + bannerText }

func sawPrint(g *fakegame.Game, arg string) func() bool {
	return func() bool {
		for _, c := range g.CommandsNamed(">") {
			if c.Arg == arg {
				return true
			}
		}
		return false
	}
}

func TestBannerEveryFrame(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen(wideScreen(40))
	startEngine(t, g, Config{Banner: bannerText})
	eventually(t, "строка в трёх кадрах", func() bool { return len(g.CommandsNamed(">")) >= 3 })
	cmds := g.CommandsNamed(">")
	for i, c := range cmds {
		if c.Arg != bannerAt("19") || c.Phase != "post" {
			t.Fatalf("команда печати = %+v, ожидалась строка %q в шаге post", c, bannerAt("19"))
		}
		if i > 0 && c.Frame != cmds[i-1].Frame+1 {
			t.Fatalf("строка пришла в кадре %d после кадра %d, ожидался каждый кадр", c.Frame, cmds[i-1].Frame)
		}
	}
}

func TestBannerFollowsScreenWidth(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen(wideScreen(40))
	startEngine(t, g, Config{Banner: bannerText})
	eventually(t, "строка у правого края экрана 40", sawPrint(g, bannerAt("19")))
	g.SetScreen(wideScreen(60))
	eventually(t, "строка у правого края экрана 60", sawPrint(g, bannerAt("39")))
	g.SetScreen(wideScreen(5))
	eventually(t, "строка с нулевого столбца на узком экране", sawPrint(g, bannerAt("0")))
}

func TestBannerBeforeAssistantText(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen(wideScreen(40))
	e := startEngine(t, g, Config{Banner: bannerText})
	text, err := Print("hi", &Point{X: 19}, "", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.Do(testContext(t), text, 0); err != nil {
		t.Fatal(err)
	}
	held := 0
	eventually(t, "текст ассистента в трёх кадрах", func() bool {
		held = 0
		for _, c := range g.CommandsNamed(">") {
			if c.Arg == "`19,0,hi" {
				held++
			}
		}
		return held >= 3
	})
	cmds := g.CommandsNamed(">")
	for i, c := range cmds {
		if c.Arg == "`19,0,hi" && (i == 0 || cmds[i-1].Arg != bannerAt("19") || cmds[i-1].Frame != c.Frame) {
			t.Fatalf("перед текстом ассистента в кадре %d нет строки версии", c.Frame)
		}
	}
}

func TestBannerAfterReconnect(t *testing.T) {
	g := fakegame.Start(t)
	startEngine(t, g, Config{Banner: bannerText})
	g.DropConn()
	n := len(g.CommandsNamed(">"))
	eventually(t, "строка после восстановления связи", func() bool { return len(g.CommandsNamed(">")) > n })
}

func TestBannerNotRepeatedInFreeze(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{Banner: bannerText})
	freezeNow(t, e, time.Minute)
	n := len(g.CommandsNamed(">"))
	if _, err := e.Do(testContext(t), Activate("potion"), 0); err != nil {
		t.Fatal(err)
	}
	if got := len(g.CommandsNamed(">")); got != n {
		t.Errorf("в заморозке строка отправлена повторно: было %d команд печати, стало %d", n, got)
	}
	e.Release()
}

func TestNoBannerByDefault(t *testing.T) {
	g := fakegame.Start(t)
	startEngine(t, g, Config{})
	from := g.Frame()
	eventually(t, "три кадра без строки", func() bool { return g.Frame() >= from+3 })
	if cmds := g.CommandsNamed(">"); len(cmds) != 0 {
		t.Errorf("без текста строки игра получила печать %+v", cmds)
	}
}

func TestBannerRejected(t *testing.T) {
	for _, text := range []string{"stonemcp @hp@", "#fff"} {
		if err := New(Config{Banner: text}).Run(context.Background()); err == nil {
			t.Errorf("движок принял строку %q, которую игра не напечатает как текст", text)
		}
	}
}

func TestBannerTakesFrameBudget(t *testing.T) {
	if got := newCommandQueue(Config{FrameBudget: 10, Banner: "x"}).budget; got != 9 {
		t.Errorf("бюджет команд со строкой = %d, ожидалось 9", got)
	}
	if got := newCommandQueue(Config{FrameBudget: 10}).budget; got != 10 {
		t.Errorf("бюджет команд без строки = %d, ожидалось 10", got)
	}
}
