package engine

// Foe содержит состояние текущего врага.
type Foe struct {
	// Text содержит исходное описание врага.
	Text string `json:"text"`
	// ID содержит идентификатор типа врага.
	ID string `json:"id"`
	// Name содержит имя врага.
	Name string `json:"name"`
	// HP содержит здоровье.
	HP int `json:"hp"`
	// MaxHP содержит максимальное здоровье.
	MaxHP int `json:"maxhp"`
	// Armor содержит броню.
	Armor int `json:"armor"`
	// Distance содержит расстояние до врага.
	Distance int `json:"distance"`
	// Count содержит количество врагов рядом.
	Count int `json:"count"`
	// State содержит код состояния врага.
	State int `json:"state"`
	// Time содержит время состояния врага в кадрах.
	Time int `json:"time"`
	// Buffs содержит усиления в формате игры.
	Buffs string `json:"buffs"`
	// Debuffs содержит ослабления в формате игры.
	Debuffs string `json:"debuffs"`
	// Element содержит стихию врага по его тегам или пусто без стихии.
	Element string `json:"element"`
}
