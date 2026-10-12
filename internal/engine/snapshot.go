package engine

import "time"

// Snapshot содержит состояние одного кадра и возраст данных.
type Snapshot struct {
	// Frame содержит номер кадра движка.
	Frame int64 `json:"frame"`
	// AgeMS содержит возраст данных в миллисекундах.
	AgeMS int64 `json:"age_ms"`
	// Time содержит игровое время локации в кадрах.
	Time int `json:"time"`
	// TotalTime содержит общее игровое время в кадрах.
	TotalTime int `json:"totaltime"`
	// Player содержит состояние персонажа.
	Player Player `json:"player"`
	// AI содержит состояние автоматики игры.
	AI AI `json:"ai"`
	// Loc содержит состояние локации.
	Loc Loc `json:"loc"`
	// Foe содержит текущего врага или nil при его отсутствии.
	Foe *Foe `json:"foe"`
	// Screen содержит размеры экрана в клетках.
	Screen Screen `json:"screen"`
	taken  time.Time
}

func (s *Snapshot) aged(now time.Time) *Snapshot {
	copy := *s
	copy.AgeMS = max(0, now.Sub(s.taken).Milliseconds())
	if s.Foe != nil {
		foe := *s.Foe
		copy.Foe = &foe
	}
	return &copy
}
