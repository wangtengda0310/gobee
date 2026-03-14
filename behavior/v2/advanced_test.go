package behavior

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandomSelector(t *testing.T) {
	ctx := make(Context)

	// Test with all failing children
	failure1 := NewAction(func(_ Context) Result { return Failure })
	failure2 := NewAction(func(_ Context) Result { return Failure })

	selector := NewRandomSelector(failure1, failure2)
	result := selector.Tick(ctx)
	assert.Equal(t, Failure, result, "RandomSelector should return Failure when all children fail")

	// Test with one success
	success := NewAction(func(_ Context) Result { return Success })
	selector = NewRandomSelector(failure1, success, failure2)
	result = selector.Tick(ctx)
	assert.Equal(t, Success, result, "RandomSelector should return Success when one child succeeds")

	// Test empty selector
	emptySelector := NewRandomSelector()
	result = emptySelector.Tick(ctx)
	assert.Equal(t, Failure, result, "Empty RandomSelector should return Failure")

	// Test AddChild
	selector = NewRandomSelector()
	selector.AddChild(success)
	result = selector.Tick(ctx)
	assert.Equal(t, Success, result, "RandomSelector with AddChild should succeed")
}

func TestRetry(t *testing.T) {
	ctx := make(Context)

	// Test retry on failure
	failCount := 0
	failingAction := NewAction(func(_ Context) Result {
		failCount++
		if failCount >= 3 {
			return Success
		}
		return Failure
	})

	retry := NewRetry(5, failingAction)

	// First attempt - fails, retry
	result := retry.Tick(ctx)
	assert.Equal(t, Running, result, "Retry should return Running on first failure")
	assert.Equal(t, 1, failCount)

	// Second attempt - fails, retry
	result = retry.Tick(ctx)
	assert.Equal(t, Running, result, "Retry should return Running on second failure")
	assert.Equal(t, 2, failCount)

	// Third attempt - succeeds
	result = retry.Tick(ctx)
	assert.Equal(t, Success, result, "Retry should return Success when child succeeds")
	assert.Equal(t, 3, failCount)

	// Test max tries exceeded
	failCount = 0
	alwaysFail := NewAction(func(_ Context) Result {
		failCount++
		return Failure
	})
	retry = NewRetry(3, alwaysFail)

	retry.Tick(ctx)
	assert.Equal(t, 1, failCount)
	retry.Tick(ctx)
	assert.Equal(t, 2, failCount)
	result = retry.Tick(ctx)
	assert.Equal(t, Failure, result, "Retry should return Failure after max tries exceeded")
	assert.Equal(t, 3, failCount)

	// Test infinite retry (-1)
	failCount = 0
	eventuallySucceed := NewAction(func(_ Context) Result {
		failCount++
		if failCount >= 5 {
			return Success
		}
		return Failure
	})
	retry = NewRetry(-1, eventuallySucceed)
	for i := 0; i < 4; i++ {
		result = retry.Tick(ctx)
		assert.Equal(t, Running, result, "Infinite retry should return Running while failing")
	}
	result = retry.Tick(ctx)
	assert.Equal(t, Success, result, "Infinite retry should succeed when child succeeds")
	assert.Equal(t, 5, failCount)

	// Test SetChild
	newAction := NewAction(func(_ Context) Result { return Success })
	retry = NewRetry(1, nil)
	retry.SetChild(newAction)
	result = retry.Tick(ctx)
	assert.Equal(t, Success, result, "Retry with SetChild should succeed")

	// Test Reset
	failCount = 0
	failTwice := NewAction(func(_ Context) Result {
		failCount++
		if failCount >= 3 {
			return Success
		}
		return Failure
	})
	retry = NewRetry(5, failTwice)
	retry.Tick(ctx)
	retry.Tick(ctx)
	retry.Reset()
	failCount = 0 // Reset counter
	result = retry.Tick(ctx)
	assert.Equal(t, Running, result, "Retry after Reset should start fresh")

	// Test child returns Running
	runningCount := 0
	runningAction := NewAction(func(_ Context) Result {
		runningCount++
		if runningCount < 3 {
			return Running
		}
		return Success
	})
	retry = NewRetry(5, runningAction)
	result = retry.Tick(ctx)
	assert.Equal(t, Running, result, "Retry should return Running when child is Running")
	result = retry.Tick(ctx)
	assert.Equal(t, Running, result)
	result = retry.Tick(ctx)
	assert.Equal(t, Success, result, "Retry should succeed when child succeeds after Running")

	// Test nil child
	retry = NewRetry(3, nil)
	result = retry.Tick(ctx)
	assert.Equal(t, Failure, result, "Retry with nil child should return Failure")
}

func TestTimeout(t *testing.T) {
	ctx := make(Context)

	// Test timeout
	action := NewAction(func(_ Context) Result { return Running })
	timeout := NewTimeout(100 * time.Millisecond, action)

	// Should return Running initially
	result := timeout.Tick(ctx)
	assert.Equal(t, Running, result, "Timeout should return Running initially")

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)
	result = timeout.Tick(ctx)
	assert.Equal(t, Failure, result, "Timeout should return Failure after duration")

	// Test with immediate success
	success := NewAction(func(_ Context) Result { return Success })
	timeout = NewTimeout(1 * time.Second, success)
	result = timeout.Tick(ctx)
	assert.Equal(t, Success, result, "Timeout should return child's Success")

	// Test with immediate failure
	failAction := NewAction(func(_ Context) Result { return Failure })
	timeout = NewTimeout(1 * time.Second, failAction)
	result = timeout.Tick(ctx)
	assert.Equal(t, Failure, result, "Timeout should return child's Failure")

	// Test SetChild
	timeout = NewTimeout(1*time.Second, nil)
	timeout.SetChild(success)
	result = timeout.Tick(ctx)
	assert.Equal(t, Success, result, "Timeout with SetChild should succeed")

	// Test Reset
	timeout = NewTimeout(50*time.Millisecond, action)
	timeout.Tick(ctx)
	time.Sleep(100 * time.Millisecond)
	timeout.Reset()
	// After reset, should work again
	result = timeout.Tick(ctx)
	assert.Equal(t, Running, result, "Timeout after Reset should work again")

	// Test nil child
	timeout = NewTimeout(1*time.Second, nil)
	result = timeout.Tick(ctx)
	assert.Equal(t, Failure, result, "Timeout with nil child should return Failure")
}

func TestDelay(t *testing.T) {
	ctx := make(Context)

	executed := false
	action := NewAction(func(_ Context) Result {
		executed = true
		return Success
	})

	delay := NewDelay(3, action)

	// First tick - delay
	result := delay.Tick(ctx)
	assert.Equal(t, Running, result, "Delay should return Running on first tick")
	assert.False(t, executed, "Action should not be executed yet")

	// Second tick - delay
	result = delay.Tick(ctx)
	assert.Equal(t, Running, result, "Delay should return Running on second tick")
	assert.False(t, executed, "Action should not be executed yet")

	// Third tick - delay
	result = delay.Tick(ctx)
	assert.Equal(t, Running, result, "Delay should return Running on third tick")
	assert.False(t, executed, "Action should not be executed yet")

	// Fourth tick - execute
	result = delay.Tick(ctx)
	assert.Equal(t, Success, result, "Delay should return Success after delay")
	assert.True(t, executed, "Action should be executed after delay")

	// Test zero delay
	executed = false
	delay = NewDelay(0, action)
	result = delay.Tick(ctx)
	assert.Equal(t, Success, result, "Zero delay should execute immediately")
	assert.True(t, executed, "Action should be executed immediately with zero delay")

	// Test SetChild
	executed = false
	newAction := NewAction(func(_ Context) Result {
		executed = true
		return Success
	})
	delay = NewDelay(1, nil)
	delay.SetChild(newAction)
	delay.Tick(ctx) // delay tick
	result = delay.Tick(ctx)
	assert.Equal(t, Success, result, "Delay with SetChild should work")
	assert.True(t, executed)

	// Test Reset
	executed = false
	delay = NewDelay(2, action)
	delay.Tick(ctx)
	delay.Reset()
	// After reset, should start delay from beginning
	executed = false
	delay.Tick(ctx)
	assert.False(t, executed, "Should be in delay phase after Reset")

	// Test child returns Running
	runningCount := 0
	runningAction := NewAction(func(_ Context) Result {
		runningCount++
		if runningCount < 2 {
			return Running
		}
		return Success
	})
	delay = NewDelay(1, runningAction)
	delay.Tick(ctx) // delay tick
	result = delay.Tick(ctx)
	assert.Equal(t, Running, result, "Delay should return Running when child is Running")
	result = delay.Tick(ctx)
	assert.Equal(t, Success, result, "Delay should succeed when child succeeds")

	// Test nil child
	delay = NewDelay(1, nil)
	result = delay.Tick(ctx)
	assert.Equal(t, Failure, result, "Delay with nil child should return Failure")
}

func TestLimiter(t *testing.T) {
	ctx := make(Context)

	callCount := 0
	action := NewAction(func(_ Context) Result {
		callCount++
		return Success
	})

	limiter := NewLimiter(3, action)

	// First call - allowed
	result := limiter.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, 1, callCount)

	// Second call - allowed
	result = limiter.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, 2, callCount)

	// Third call - allowed
	result = limiter.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, 3, callCount)

	// Fourth call - blocked
	result = limiter.Tick(ctx)
	assert.Equal(t, Failure, result, "Limiter should block calls after maxCalls")
	assert.Equal(t, 3, callCount, "Call count should not increase after limit")

	// Test Reset
	limiter.Reset()
	result = limiter.Tick(ctx)
	assert.Equal(t, Success, result, "Limiter should allow calls after Reset")
	assert.Equal(t, 4, callCount)

	// Test unlimited (-1)
	callCount = 0
	limiter = NewLimiter(-1, action)
	for i := 0; i < 10; i++ {
		result = limiter.Tick(ctx)
		assert.Equal(t, Success, result, "Unlimited limiter should always allow")
	}
	assert.Equal(t, 10, callCount, "Unlimited limiter should count all calls")

	// Test SetChild
	callCount = 0
	newAction := NewAction(func(_ Context) Result {
		callCount++
		return Success
	})
	limiter = NewLimiter(1, nil)
	limiter.SetChild(newAction)
	result = limiter.Tick(ctx)
	assert.Equal(t, Success, result, "Limiter with SetChild should work")
	assert.Equal(t, 1, callCount)

	// Test only success counts
	failCount := 0
	failAction := NewAction(func(_ Context) Result {
		failCount++
		return Failure
	})
	limiter = NewLimiter(3, failAction)
	limiter.Tick(ctx)
	limiter.Tick(ctx)
	limiter.Tick(ctx)
	limiter.Tick(ctx)
	assert.Equal(t, 4, failCount, "Failure should not count towards limit")

	// Test nil child
	limiter = NewLimiter(1, nil)
	result = limiter.Tick(ctx)
	assert.Equal(t, Failure, result, "Limiter with nil child should return Failure")
}
