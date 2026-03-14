package main

import (
	"fmt"

	"github.com/wangtengda/gobee/behavior/v2"
)

// Example demonstrates how to use the behavior tree library.
func Example() {
	// Create a context to pass data between nodes
	ctx := make(behavior.Context)
	ctx["counter"] = 0
	ctx["target"] = 3

	// Create a condition that checks if counter has reached target
	isCounterReached := behavior.NewCondition(func(ctx behavior.Context) bool {
		counter := ctx["counter"].(int)
		target := ctx["target"].(int)
		return counter >= target
	})

	// Create an action that increments the counter
	incrementCounter := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
		counter := ctx["counter"].(int)
		counter++
		ctx["counter"] = counter
		fmt.Printf("Counter: %d\n", counter)
		return behavior.Success
	})

	// Create a repeater that repeats the increment action until the condition is met
	repeater := behavior.NewRepeater(-1, incrementCounter) // -1 for infinite repeats

	// Create a selector that tries the condition first, then the repeater
	tree := behavior.NewSelector(
		isCounterReached,
		repeater,
	)

	// Tick the tree until it succeeds
	result := behavior.Running
	for result == behavior.Running {
		result = tree.Tick(ctx)
		fmt.Printf("Tree result: %s\n", result)
	}

	fmt.Println("Behavior tree execution completed")
}
