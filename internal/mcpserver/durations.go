package mcpserver

import (
	"fmt"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

func holdFor(s float64) (time.Duration, error) {
	if err := secondsIn(s, engine.MaxHold.Seconds(), "удержание"); err != nil {
		return 0, err
	}
	return seconds(s), nil
}

func secondsIn(v, limit float64, what string) error {
	if v < 0 || v > limit {
		return fmt.Errorf("%s должно быть от 0 до %v секунд", what, limit)
	}
	return nil
}

func seconds(s float64) time.Duration { return time.Duration(s * float64(time.Second)) }
