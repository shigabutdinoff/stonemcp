package engine

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	leftHand = 1 << iota
	rightHand
	bothHands = leftHand | rightHand
)

const (
	printMark     = "`"
	brewSeparator = " + "
)

type commandSpec struct {
	slot    func(Command) string
	hands   int
	blankOK bool
	arg     func(Command) error
}

var commandSpecs = map[string]commandSpec{
	"equip":    {slot: nameSlot, hands: bothHands, arg: wholeItem},
	"equipL":   {slot: nameSlot, hands: leftHand, arg: wholeItem},
	"equipR":   {slot: nameSlot, hands: rightHand, arg: wholeItem},
	"loadout":  {slot: nameSlot, hands: bothHands},
	"activate": {},
	"brew":     {arg: brewIngredients},
	">":        {slot: printSlot, blankOK: true, arg: printTarget},
}

func nameSlot(c Command) string { return c.Name }

func printSlot(c Command) string {
	if pos, ok := printPosition(c.Arg); ok {
		return "print:" + pos
	}
	return "print:default"
}

func printPosition(arg string) (string, bool) {
	rest, ok := strings.CutPrefix(arg, printMark)
	if !ok {
		return "", false
	}
	parts := strings.SplitN(rest, ",", 3)
	if len(parts) < 3 {
		return "", false
	}
	x, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", false
	}
	y, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", false
	}
	return strconv.Itoa(x) + "," + strconv.Itoa(y), true
}

func wholeItem(c Command) error {
	if strings.Contains(c.Arg, "//") {
		return fmt.Errorf("игра обрежет аргумент команды %q на первом //", c.Name)
	}
	return nil
}

func printTarget(c Command) error {
	if strings.HasPrefix(c.Arg, printMark) {
		if _, ok := printPosition(c.Arg); !ok {
			return fmt.Errorf("текст с обратной кавычки требует позиции вида `x,y,")
		}
	}
	return nil
}

func brewIngredients(c Command) error {
	for ingredient := range strings.SplitSeq(c.Arg, brewSeparator) {
		if !slices.Contains(Ingredients, ingredient) {
			return fmt.Errorf("недопустимый ингредиент зелья %q", ingredient)
		}
	}
	return nil
}
