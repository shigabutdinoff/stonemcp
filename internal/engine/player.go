package engine

// Player содержит здоровье, броню, позицию и эффекты персонажа.
type Player struct {
	// HP содержит здоровье.
	HP int `json:"hp"`
	// MaxHP содержит максимальное здоровье.
	MaxHP int `json:"maxhp"`
	// Armor содержит броню.
	Armor int `json:"armor"`
	// MaxArmor содержит максимальную броню.
	MaxArmor int `json:"maxarmor"`
	// X содержит координату по горизонтали.
	X int `json:"x"`
	// Y содержит координату по вертикали.
	Y int `json:"y"`
	// Z содержит глубину позиции.
	Z int `json:"z"`
	// Buffs содержит усиления в формате игры.
	Buffs string `json:"buffs"`
	// Debuffs содержит ослабления в формате игры.
	Debuffs string `json:"debuffs"`
}

// AI содержит признаки работы автоматики игры.
type AI struct {
	// Enabled сообщает, включена ли автоматика.
	Enabled bool `json:"enabled"`
	// Paused сообщает, приостановлена ли автоматика.
	Paused bool `json:"paused"`
	// Idle сообщает, бездействует ли персонаж.
	Idle bool `json:"idle"`
	// Walking сообщает, идёт ли персонаж.
	Walking bool `json:"walking"`
}
