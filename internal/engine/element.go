package engine

import "strings"

var elements = map[string]string{
	"Stone": "Stone", "stone": "Stone", "Ice": "Ice", "Fire": "Fire",
	"Poison": "Poison", "Vigor": "Vigor", "AEther": "AEther", "Air": "Air",
}

func elementOf(foe, id, name string) string {
	words := strings.Fields(strings.TrimPrefix(foe, id+" "+name))
	for i := len(words) - 1; i >= 0; i-- {
		if element, ok := elements[words[i]]; ok {
			return element
		}
	}
	return ""
}

func foeElement(s *Snapshot) string {
	if s.Foe == nil {
		return ""
	}
	return s.Foe.Element
}
