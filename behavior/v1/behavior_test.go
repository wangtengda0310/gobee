package behavior

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResultString(t *testing.T) {
	assert.Equal(t, "Success", Success.String(), "Success should stringify to Success")
	assert.Equal(t, "Failure", Failure.String(), "Failure should stringify to Failure")
	assert.Equal(t, "Running", Running.String(), "Running should stringify to Running")
	assert.Equal(t, "Unknown", Result(99).String(), "Unknown result should stringify to Unknown")
}

func TestAction(t *testing.T) {
	// Test successful action
	successAction := Action(func(ctx Context) Result {
		return Success
	})
	result := successAction(make(Context))
	assert.Equal(t, Success, result, "Action should return Success when function returns Success")

	// Test failure action
	failureAction := Action(func(ctx Context) Result {
		return Failure
	})
	result = failureAction(make(Context))
	assert.Equal(t, Failure, result, "Action should return Failure when function returns Failure")

	// Test running action
	runningAction := Action(func(ctx Context) Result {
		return Running
	})
	result = runningAction(make(Context))
	assert.Equal(t, Running, result, "Action should return Running when function returns Running")

	// Test action with context
	ctx := make(Context)
	ctx["count"] = 0
	counterAction := Action(func(ctx Context) Result {
		count := ctx["count"].(int)
		ctx["count"] = count + 1
		return Success
	})
	result = counterAction(ctx)
	assert.Equal(t, Success, result, "Counter action should return Success")
	assert.Equal(t, 1, ctx["count"], "Counter should be incremented")
}

func TestCondition(t *testing.T) {
	// Test true condition
	trueCondition := Condition(func(ctx Context) bool {
		return true
	})
	result := trueCondition(make(Context))
	assert.Equal(t, Success, result, "Condition should return Success when function returns true")

	// Test false condition
	falseCondition := Condition(func(ctx Context) bool {
		return false
	})
	result = falseCondition(make(Context))
	assert.Equal(t, Failure, result, "Condition should return Failure when function returns false")

	// Test condition with context
	ctx := make(Context)
	ctx["value"] = 5
	valueCondition := Condition(func(ctx Context) bool {
		return ctx["value"].(int) > 3
	})
	result = valueCondition(ctx)
	assert.Equal(t, Success, result, "Value condition should return Success when condition is met")
}

func TestSequence(t *testing.T) {
	// Test all successes
	success1 := Action(func(ctx Context) Result { return Success })
	success2 := Action(func(ctx Context) Result { return Success })
	sequence := Sequence(success1, success2)
	result := sequence(make(Context))
	assert.Equal(t, Success, result, "Sequence should return Success when all children succeed")

	// Test early failure
	failure := Action(func(ctx Context) Result { return Failure })
	afterFailure := Action(func(ctx Context) Result {
		t.Error("After failure should not be executed")
		return Success
	})
	sequence = Sequence(success1, failure, afterFailure)
	result = sequence(make(Context))
	assert.Equal(t, Failure, result, "Sequence should return Failure when first child fails")

	// Test running
	running := Action(func(ctx Context) Result { return Running })
	afterRunning := Action(func(ctx Context) Result {
		t.Error("After running should not be executed")
		return Success
	})
	sequence = Sequence(success1, running, afterRunning)
	result = sequence(make(Context))
	assert.Equal(t, Running, result, "Sequence should return Running when child returns Running")

	// Test empty sequence
	emptySequence := Sequence()
	result = emptySequence(make(Context))
	assert.Equal(t, Success, result, "Empty sequence should return Success")
}

func TestSelector(t *testing.T) {
	// Test all failures
	failure1 := Action(func(ctx Context) Result { return Failure })
	failure2 := Action(func(ctx Context) Result { return Failure })
	selector := Selector(failure1, failure2)
	result := selector(make(Context))
	assert.Equal(t, Failure, result, "Selector should return Failure when all children fail")

	// Test early success
	success := Action(func(ctx Context) Result { return Success })
	afterSuccess := Action(func(ctx Context) Result {
		t.Error("After success should not be executed")
		return Failure
	})
	selector = Selector(failure1, success, afterSuccess)
	result = selector(make(Context))
	assert.Equal(t, Success, result, "Selector should return Success when first child succeeds")

	// Test running
	running := Action(func(ctx Context) Result { return Running })
	afterRunning := Action(func(ctx Context) Result {
		t.Error("After running should not be executed")
		return Failure
	})
	selector = Selector(failure1, running, afterRunning)
	result = selector(make(Context))
	assert.Equal(t, Running, result, "Selector should return Running when child returns Running")

	// Test empty selector
	emptySelector := Selector()
	result = emptySelector(make(Context))
	assert.Equal(t, Failure, result, "Empty selector should return Failure")
}

func TestParallel(t *testing.T) {
	// Test success policy
	success1 := Action(func(ctx Context) Result { return Success })
	success2 := Action(func(ctx Context) Result { return Success })
	// Require 2 successes to succeed
	parallel := Parallel(2, 1, success1, success2)
	result := parallel(make(Context))
	assert.Equal(t, Success, result, "Parallel should return Success when success policy is met")

	// Test failure policy
	failure1 := Action(func(ctx Context) Result { return Failure })
	// Require 1 failure to fail
	parallel = Parallel(2, 1, success1, failure1)
	result = parallel(make(Context))
	assert.Equal(t, Failure, result, "Parallel should return Failure when failure policy is met")

	// Test running
	running1 := Action(func(ctx Context) Result { return Running })
	parallel = Parallel(2, 1, success1, running1)
	result = parallel(make(Context))
	assert.Equal(t, Running, result, "Parallel should return Running when any child returns Running")

	// Test mixed results without meeting policies
	parallel = Parallel(2, 2, success1, failure1)
	result = parallel(make(Context))
	assert.Equal(t, Failure, result, "Parallel should return Failure when no policies are met")

	// Test empty parallel
	emptyParallel := Parallel(1, 1)
	result = emptyParallel(make(Context))
	assert.Equal(t, Success, result, "Empty parallel should return Success")
}

func TestInverter(t *testing.T) {
	// Test inverting success
	success := Action(func(ctx Context) Result { return Success })
	inverter := Inverter(success)
	result := inverter(make(Context))
	assert.Equal(t, Failure, result, "Success result should be inverted to Failure")

	// Test inverting failure
	failure := Action(func(ctx Context) Result { return Failure })
	inverter = Inverter(failure)
	result = inverter(make(Context))
	assert.Equal(t, Success, result, "Failure result should be inverted to Success")

	// Test inverting running
	running := Action(func(ctx Context) Result { return Running })
	inverter = Inverter(running)
	result = inverter(make(Context))
	assert.Equal(t, Running, result, "Running result should remain Running after inversion")
}

func TestRepeater(t *testing.T) {
	// Test repeating 2 times
	ctx := make(Context)
	counter := 0
	counterAction := Action(func(ctx Context) Result {
		counter++
		return Success
	})
	repeater := Repeater(2, counterAction)

	// First tick - should return Running
	result := repeater(ctx)
	assert.Equal(t, Running, result, "First tick of repeater should return Running")
	assert.Equal(t, 1, counter, "Counter should be incremented once after first tick")

	// Second tick - should return Success (completed 2 repetitions)
	result = repeater(ctx)
	assert.Equal(t, Success, result, "Repeater should return Success after completing all repetitions")
	assert.Equal(t, 2, counter, "Counter should be incremented twice after completing repetitions")

	// Test repeating 0 times (should succeed immediately)
	counter = 0
	repeater = Repeater(0, counterAction)
	result = repeater(make(Context))
	assert.Equal(t, Success, result, "Repeater with 0 times should return Success immediately")
	assert.Equal(t, 0, counter, "Counter should not be incremented when repeater has 0 times")

	// Test running child
	ctx = make(Context)
	runningAction := Action(func(ctx Context) Result { return Running })
	repeater = Repeater(2, runningAction)
	result = repeater(ctx)
	assert.Equal(t, Running, result, "Repeater should return Running when child returns Running")
}

func TestUntilSuccess(t *testing.T) {
	// Test success on first try
	success := Action(func(ctx Context) Result { return Success })
	untilSuccess := UntilSuccess(success)
	result := untilSuccess(make(Context))
	assert.Equal(t, Success, result, "UntilSuccess should return Success when child succeeds on first try")

	// Test failure then success
	ctx := make(Context)
	counter := 0
	failingThenSuccess := Action(func(ctx Context) Result {
		counter++
		if counter < 2 {
			return Failure
		}
		return Success
	})
	untilSuccess = UntilSuccess(failingThenSuccess)

	// First tick - should return Running (child failed)
	result = untilSuccess(ctx)
	assert.Equal(t, Running, result, "First tick with failing child should return Running")
	assert.Equal(t, 1, counter, "Counter should be incremented once after first tick")

	// Second tick - should return Success (child succeeded)
	result = untilSuccess(ctx)
	assert.Equal(t, Success, result, "Second tick with succeeding child should return Success")
	assert.Equal(t, 2, counter, "Counter should be incremented twice after second tick")

	// Test running child
	running := Action(func(ctx Context) Result { return Running })
	untilSuccess = UntilSuccess(running)
	result = untilSuccess(make(Context))
	assert.Equal(t, Running, result, "UntilSuccess should return Running when child returns Running")
}

func TestUntilFailure(t *testing.T) {
	// Test failure on first try
	failure := Action(func(ctx Context) Result { return Failure })
	untilFailure := UntilFailure(failure)
	result := untilFailure(make(Context))
	assert.Equal(t, Success, result, "UntilFailure should return Success when child fails on first try")

	// Test success then failure
	ctx := make(Context)
	counter := 0
	succeedingThenFailure := Action(func(ctx Context) Result {
		counter++
		if counter < 2 {
			return Success
		}
		return Failure
	})
	untilFailure = UntilFailure(succeedingThenFailure)

	// First tick - should return Running (child succeeded)
	result = untilFailure(ctx)
	assert.Equal(t, Running, result, "First tick with succeeding child should return Running")
	assert.Equal(t, 1, counter, "Counter should be incremented once after first tick")

	// Second tick - should return Success (child failed)
	result = untilFailure(ctx)
	assert.Equal(t, Success, result, "Second tick with failing child should return Success")
	assert.Equal(t, 2, counter, "Counter should be incremented twice after second tick")

	// Test running child
	running := Action(func(ctx Context) Result { return Running })
	untilFailure = UntilFailure(running)
	result = untilFailure(make(Context))
	assert.Equal(t, Running, result, "UntilFailure should return Running when child returns Running")
}

func TestExample(t *testing.T) {
	// Test that Example.Run() executes without panic
	example := &Example{}
	// This is primarily a smoke test to ensure the example code works
	assert.NotPanics(t, func() {
		example.Run()
	}, "Example.Run() should not panic")
}

func TestDemoBehaviorTree(t *testing.T) {
	// Test that DemoBehaviorTree executes without panic
	assert.NotPanics(t, func() {
		DemoBehaviorTree()
	}, "DemoBehaviorTree() should not panic")
}

// === 新节点测试 ===

func TestFilter(t *testing.T) {
	shoot := Action(func(ctx Context) Result { return Success })

	// Test condition satisfied
	ctx := make(Context)
	ctx["hasAmmo"] = true
	filter := Filter(func(ctx Context) bool {
		return ctx["hasAmmo"].(bool)
	}, shoot)
	result := filter(ctx)
	assert.Equal(t, Success, result, "Filter should return Success when condition is met and child succeeds")

	// Test condition not satisfied
	ctx["hasAmmo"] = false
	result = filter(ctx)
	assert.Equal(t, Failure, result, "Filter should return Failure when condition is not met")
}

func TestActiveSelector(t *testing.T) {
	// Test all failures
	failure1 := Action(func(ctx Context) Result { return Failure })
	failure2 := Action(func(ctx Context) Result { return Failure })
	selector := ActiveSelector(failure1, failure2)
	result := selector(make(Context))
	assert.Equal(t, Failure, result, "ActiveSelector should return Failure when all children fail")

	// Test first success
	success := Action(func(ctx Context) Result { return Success })
	afterSuccess := Action(func(ctx Context) Result {
		t.Error("After success should not be executed")
		return Failure
	})
	selector = ActiveSelector(success, afterSuccess)
	result = selector(make(Context))
	assert.Equal(t, Success, result, "ActiveSelector should return Success when first child succeeds")

	// Test running
	running := Action(func(ctx Context) Result { return Running })
	afterRunning := Action(func(ctx Context) Result {
		t.Error("After running should not be executed")
		return Failure
	})
	selector = ActiveSelector(running, afterRunning)
	result = selector(make(Context))
	assert.Equal(t, Running, result, "ActiveSelector should return Running when child returns Running")

	// Test empty active selector
	emptySelector := ActiveSelector()
	result = emptySelector(make(Context))
	assert.Equal(t, Failure, result, "Empty ActiveSelector should return Failure")
}

func TestMonitor(t *testing.T) {
	attack := Action(func(ctx Context) Result { return Success })

	// Test condition met
	ctx := make(Context)
	ctx["hasEnemy"] = true
	monitor := Monitor(func(ctx Context) bool {
		return ctx["hasEnemy"].(bool)
	}, attack)
	result := monitor(ctx)
	assert.Equal(t, Success, result, "Monitor should return child's result when condition is met")

	// Test condition not met
	ctx["hasEnemy"] = false
	result = monitor(ctx)
	assert.Equal(t, Failure, result, "Monitor should return Failure when condition is not met")
}

func TestRepeat(t *testing.T) {
	ctx := make(Context)
	counter := 0

	// Test child succeeds - should keep running
	successAction := Action(func(ctx Context) Result {
		counter++
		return Success
	})
	repeat := Repeat(successAction)
	result := repeat(ctx)
	assert.Equal(t, Running, result, "Repeat should return Running when child succeeds")
	assert.Equal(t, 1, counter, "Counter should be incremented after first tick")

	// Second tick
	result = repeat(ctx)
	assert.Equal(t, Running, result, "Repeat should keep returning Running")
	assert.Equal(t, 2, counter, "Counter should be incremented after second tick")

	// Test child running
	runningAction := Action(func(ctx Context) Result { return Running })
	repeat = Repeat(runningAction)
	result = repeat(ctx)
	assert.Equal(t, Running, result, "Repeat should return Running when child is running")

	// Test child failure
	counter = 0
	failureAction := Action(func(ctx Context) Result {
		counter++
		return Failure
	})
	repeat = Repeat(failureAction)
	result = repeat(ctx)
	assert.Equal(t, Running, result, "Repeat should return Running when child fails")
	assert.Equal(t, 1, counter, "Counter should be incremented even when child fails")
}
