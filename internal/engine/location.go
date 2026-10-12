package engine

// Loc содержит текущую локацию и признаки начала и петли забега.
type Loc struct {
	// ID содержит идентификатор локации.
	ID string `json:"id"`
	// Name содержит название локации.
	Name string `json:"name"`
	// Stars содержит сложность локации.
	Stars int `json:"stars"`
	// Begin сообщает, начался ли забег в этом кадре.
	Begin bool `json:"begin"`
	// Loop сообщает, началась ли петля в этом кадре.
	Loop bool `json:"loop"`
}

// Screen содержит размеры экрана в клетках.
type Screen struct {
	// W содержит ширину.
	W int `json:"w"`
	// H содержит высоту.
	H int `json:"h"`
}
