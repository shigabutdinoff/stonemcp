package engine_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

func ExampleNew() {
	ctx, cancel := context.WithCancel(context.Background())
	eng := engine.New(engine.Config{Addr: engine.DefaultAddr})
	var running sync.WaitGroup
	running.Go(func() {
		if err := eng.Run(ctx); err != nil {
			fmt.Println(err)
		}
	})
	fmt.Println(eng.Status().Connected)
	cancel()
	running.Wait()
}

func ExampleEngine_Evaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	eng := start(ctx)
	values, err := eng.Evaluate(ctx, []engine.Query{{Name: "hp"}, {Name: "maxhp"}})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(values[0], values[1])
}

func ExampleEngine_Do() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	eng := start(ctx)
	result, err := eng.Do(ctx, engine.Equip("sword", 0), 0)
	var gameErr *engine.GameError
	switch {
	case errors.As(err, &gameErr):
		fmt.Println(gameErr.Code)
	case err != nil:
		fmt.Println(err)
	default:
		fmt.Println(result.Result, result.Frozen)
	}
}

func ExampleEngine_WaitFor() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eng := start(ctx)
	conditions := []engine.Condition{
		{Kind: engine.CondFoePresent},
		{Kind: engine.CondHPBelow, Value: 0.5},
	}
	result, err := eng.WaitFor(ctx, conditions, 30*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result.Reason)
}

func start(ctx context.Context) *engine.Engine {
	eng := engine.New(engine.Config{})
	go func() {
		if err := eng.Run(ctx); err != nil {
			fmt.Println(err)
		}
	}()
	return eng
}
