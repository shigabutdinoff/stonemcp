package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Equip выбирает экипировку обеих рук с заданным удержанием.
func Equip(item string, hold time.Duration) Command {
	return Command{Name: "equip", Arg: item, Hold: hold}
}

// EquipLeft выбирает экипировку левой руки с заданным удержанием.
func EquipLeft(item string, hold time.Duration) Command {
	return Command{Name: "equipL", Arg: item, Hold: hold}
}

// EquipRight выбирает экипировку правой руки с заданным удержанием.
func EquipRight(item string, hold time.Duration) Command {
	return Command{Name: "equipR", Arg: item, Hold: hold}
}

// Loadout выбирает набор экипировки с заданным удержанием.
func Loadout(number int, hold time.Duration) Command {
	return Command{Name: "loadout", Arg: strconv.Itoa(number), Hold: hold}
}

// Activate активирует способность или зелье без удержания.
func Activate(target string) Command { return Command{Name: "activate", Arg: target} }

// Ingredients перечисляет ингредиенты, которые принимает Brew.
var Ingredients = []string{"stone", "wood", "tar", "bronze"}

// Brew смешивает зелье из переданных ингредиентов без удержания.
func Brew(ingredients []string) Command {
	return Command{Name: "brew", Arg: strings.Join(ingredients, brewSeparator)}
}

// Point задаёт позицию текста на экране.
type Point struct {
	// X задаёт столбец текста.
	X int
	// Y задаёт строку текста.
	Y int
}

// Print задаёт текст с удержанием в позиции at или в месте по умолчанию.
// Цвет требует позиции, текст без позиции не начинается с кавычки.
// С позицией и пустым цветом текст не начинается с решётки.
func Print(text string, at *Point, color string, hold time.Duration) (Command, error) {
	if at == nil && color != "" {
		return Command{}, fmt.Errorf("цвет требует позиции текста")
	}
	if at == nil && strings.HasPrefix(text, printMark) {
		return Command{}, fmt.Errorf("текст без позиции не может начинаться с обратной кавычки")
	}
	if at != nil && color == "" && strings.HasPrefix(text, "#") {
		return Command{}, fmt.Errorf("текст с позицией не может начинаться с решётки, игра примет его за цвет")
	}
	if strings.ContainsAny(color, ","+printMark) {
		return Command{}, fmt.Errorf("цвет не может содержать запятую или обратную кавычку")
	}
	arg := text
	if at != nil {
		prefix := fmt.Sprintf("%s%d,%d,", printMark, at.X, at.Y)
		if color != "" {
			prefix += "#" + strings.TrimPrefix(color, "#") + ","
		}
		arg = prefix + text
	}
	return Command{Name: ">", Arg: arg, Hold: hold}, nil
}
