package engine

import (
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

var baseVars = []string{
	"time", "totaltime", "hp", "maxhp", "armor", "maxarmor",
	"pos.x", "pos.y", "pos.z",
	"ai.enabled", "ai.paused", "ai.idle", "ai.walking",
	"buffs.string", "debuffs.string",
	"loc.id", "loc.name", "loc.stars", "loc.begin", "loc.loop",
	"foe", "foe.id", "foe.name", "foe.hp", "foe.maxhp", "foe.armor",
	"foe.distance", "foe.count", "foe.state", "foe.time",
	"foe.buffs.string", "foe.debuffs.string", "screen.w", "screen.h",
}

var baseRequests = func() []mc.Request {
	reqs := make([]mc.Request, len(baseVars))
	for i, name := range baseVars {
		reqs[i] = mc.Var(name)
	}
	return reqs
}()

func buildSnapshot(values []mc.Value, frame int64, now time.Time) *Snapshot {
	v := make(map[string]mc.Value, len(baseVars))
	for i, name := range baseVars {
		if i < len(values) {
			v[name] = values[i]
		}
	}
	num := func(name string) int { n, _ := v[name].Int(); return n }
	flag := func(name string) bool { b, _ := v[name].Bool(); return b }
	text := func(name string) string {
		if v[name].IsError() {
			return ""
		}
		return v[name].String()
	}
	s := &Snapshot{
		Frame: frame, Time: num("time"), TotalTime: num("totaltime"), taken: now,
		Player: Player{
			HP: num("hp"), MaxHP: num("maxhp"), Armor: num("armor"), MaxArmor: num("maxarmor"),
			X: num("pos.x"), Y: num("pos.y"), Z: num("pos.z"),
			Buffs: text("buffs.string"), Debuffs: text("debuffs.string"),
		},
		AI:     AI{Enabled: flag("ai.enabled"), Paused: flag("ai.paused"), Idle: flag("ai.idle"), Walking: flag("ai.walking")},
		Loc:    Loc{ID: text("loc.id"), Name: text("loc.name"), Stars: num("loc.stars"), Begin: flag("loc.begin"), Loop: flag("loc.loop")},
		Screen: Screen{W: num("screen.w"), H: num("screen.h")},
	}
	if text("foe") != "" {
		s.Foe = &Foe{
			Text: text("foe"), ID: text("foe.id"), Name: text("foe.name"),
			HP: num("foe.hp"), MaxHP: num("foe.maxhp"), Armor: num("foe.armor"),
			Distance: num("foe.distance"), Count: num("foe.count"), State: num("foe.state"), Time: num("foe.time"),
			Buffs: text("foe.buffs.string"), Debuffs: text("foe.debuffs.string"),
		}
		s.Foe.Element = elementOf(s.Foe.Text, s.Foe.ID, s.Foe.Name)
	}
	return s
}
