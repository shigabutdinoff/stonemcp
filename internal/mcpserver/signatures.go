package mcpserver

import (
	"github.com/shigabutdinoff/stonemcp/internal/engine"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

type kind int

const (
	anyArg kind = iota
	intArg
	strArg
)

type signature struct {
	usage    string
	args     []kind
	optional []kind
	check    func(arg any) error
	guard    func(args []mc.Arg) *engine.Guard
}

const (
	indexUsage = "номер не меньше 0"
	plainUsage = "простое имя переменной скрипта Mind Stone без точек"
)

var functions = map[string]signature{
	"app.GetStateNumber":         {usage: "app.GetStateNumber(state string), имя состояния игры, например Playing", args: []kind{strArg}, check: stateName},
	"event.GetObjectiveId":       {usage: "event.GetObjectiveId(index int), " + indexUsage, args: []kind{intArg}, check: notNegative},
	"event.GetObjectiveProgress": {usage: "event.GetObjectiveProgress(index int), " + indexUsage, args: []kind{intArg}, check: notNegative},
	"event.GetObjectiveGoal":     {usage: "event.GetObjectiveGoal(index int), " + indexUsage, args: []kind{intArg}, check: notNegative},
	"foe.GetCount":               {usage: "foe.GetCount(distance int)", args: []kind{intArg}},
	"buffs.GetCount":             {usage: "buffs.GetCount(id string)", args: []kind{strArg}},
	"buffs.GetTime":              {usage: "buffs.GetTime(id string)", args: []kind{strArg}},
	"debuffs.GetCount":           {usage: "debuffs.GetCount(id string)", args: []kind{strArg}},
	"debuffs.GetTime":            {usage: "debuffs.GetTime(id string)", args: []kind{strArg}},
	"foe.buffs.GetCount":         {usage: "foe.buffs.GetCount(id string)", args: []kind{strArg}},
	"foe.buffs.GetTime":          {usage: "foe.buffs.GetTime(id string)", args: []kind{strArg}},
	"foe.debuffs.GetCount":       {usage: "foe.debuffs.GetCount(id string)", args: []kind{strArg}},
	"foe.debuffs.GetTime":        {usage: "foe.debuffs.GetTime(id string)", args: []kind{strArg}},
	"screen.FromWorldX":          {usage: "screen.FromWorldX(x int)", args: []kind{intArg}},
	"screen.FromWorldZ":          {usage: "screen.FromWorldZ(z int)", args: []kind{intArg}},
	"screen.ToWorldX":            {usage: "screen.ToWorldX(x int)", args: []kind{intArg}},
	"screen.ToWorldZ":            {usage: "screen.ToWorldZ(z int)", args: []kind{intArg}},
	"draw.GetSymbol":             {usage: "draw.GetSymbol(x int, y int)", args: []kind{intArg, intArg}},
	"item.GetCooldown":           {usage: "item.GetCooldown(id string)", args: []kind{strArg}},
	"item.GetCount":              {usage: "item.GetCount(id string)", args: []kind{strArg}},
	"item.CanActivate":           {usage: "item.CanActivate() или item.CanActivate(ability string)", optional: []kind{strArg}},
	"item.GetLoadoutL":           {usage: "item.GetLoadoutL(number int)", args: []kind{intArg}},
	"item.GetLoadoutR":           {usage: "item.GetLoadoutR(number int)", args: []kind{intArg}},
	"item.GetTreasureCount":      {usage: "item.GetTreasureCount()"},
	"item.GetTreasureLimit":      {usage: "item.GetTreasureLimit()"},
	"summon.GetId":               {usage: "summon.GetId() или summon.GetId(index int)", optional: []kind{intArg}},
	"summon.GetName":             {usage: "summon.GetName() или summon.GetName(index int)", optional: []kind{intArg}},
	"summon.GetState":            {usage: "summon.GetState() или summon.GetState(index int)", optional: []kind{intArg}},
	"summon.GetTime":             {usage: "summon.GetTime() или summon.GetTime(index int)", optional: []kind{intArg}},
	"summon.GetVar":              {usage: `summon.GetVar("ignition") или summon.GetVar("ignition", index int)`, args: []kind{strArg}, optional: []kind{intArg}, check: summonVar, guard: onCinderwisp},
	"player.GetNextLegendName":   {usage: "player.GetNextLegendName()"},
	"var.get":                    {usage: "var.get(name), " + plainUsage, args: []kind{anyArg}, check: plainName},
	"var.has":                    {usage: "var.has(name), " + plainUsage, args: []kind{anyArg}, check: plainName},
}

var changing = []string{
	"loc.Leave", "loc.Pause", "var.set", "draw.Clear", "draw.Player", "draw.Bg",
	"draw.Box", "screen.Next", "screen.Previous", "screen.ResetOffset", "player.ShowScaredFace",
}
