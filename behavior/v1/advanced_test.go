package behavior

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandomSelector(t *testing.T) {
	ctx := make(Context)

	// Test with all failing children
	failure1 := Action(func(ctx Context) Result { return Failure })
	failure2 := Action(func(ctx Context) Result { return Failure })

	selector := RandomSelector(failure1, failure2)
	result := selector(ctx)
	assert.Equal(t, Failure, result, "RandomSelector should return Failure when all children fail")

	// Test with one success
	success := Action(func(ctx Context) Result { return Success })
	selector = RandomSelector(failure1, success, failure2)
	result = selector(ctx)
	assert.Equal(t, Success, result, "RandomSelector should return Success when one child succeeds")

	// Test empty selector
	emptySelector := RandomSelector()
	result = emptySelector(ctx)
	assert.Equal(t, Failure, result, "Empty RandomSelector should return Failure")
}

func TestRetry(t *testing.T) {
	ctx := make(Context)

	// Test retry on failure
	failCount := 0
	failingAction := Action(func(ctx Context) Result {
		failCount++
		return Failure
	})

	retry := Retry(3, failingAction)

	// First attempt
	result := retry(ctx)
	assert.Equal(t, Running, result, "Retry should return Running on first failure")
	assert.Equal(t, 1, failCount)

	// Second attempt
	result = retry(ctx)
	assert.Equal(t, Running, result, "Retry should return Running on second failure")
	assert.Equal(t, 2, failCount)

	// Third attempt - max tries exceeded
	result = retry(ctx)
	assert.Equal(t, Failure, result, "Retry should return Failure after max tries exceeded")
	assert.Equal(t, 3, failCount)

	// Test with success
	successCount := 0
	successAction := Action(func(ctx Context) Result {
		successCount++
		if successCount < 2 {
			return Failure
		}
		return Success
	})

	retry = Retry(5, successAction)
	result = retry(ctx)
	assert.Equal(t, Running, result)

	result = retry(ctx)
	assert.Equal(t, Success, result, "Retry should return Success when child succeeds")

	// Test infinite retry (-1)
	infiniteCount := 0
	infiniteAction := Action(func(ctx Context) Result {
		infiniteCount++
		if infiniteCount >= 10 {
			return Success
		}
		return Failure
	})

	retry = Retry(-1, infiniteAction)
	for i := 0; i < 9; i++ {
		result = retry(ctx)
		assert.Equal(t, Running, result, "Infinite retry should return Running while failing")
	}
	result = retry(ctx)
	assert.Equal(t, Success, result, "Infinite retry should succeed when child succeeds")
	assert.Equal(t, 10, infiniteCount)

	// Test child returns Running
	runningCount := 0
	runningAction := Action(func(ctx Context) Result {
		runningCount++
		if runningCount < 3 {
			return Running
		}
		return Success
	})

	retry = Retry(5, runningAction)
	result = retry(ctx)
	assert.Equal(t, Running, result, "Retry should return Running when child is Running")
	result = retry(ctx)
	assert.Equal(t, Running, result)
	result = retry(ctx)
	assert.Equal(t, Success, result, "Retry should succeed when child succeeds after Running")
}

func TestTimeout(t *testing.T) {
	ctx := make(Context)

	// Test timeout
	action := Action(func(ctx Context) Result { return Running })
	timeout := Timeout(100*time.Millisecond, action)

	// Should return Running initially
	result := timeout(ctx)
	assert.Equal(t, Running, result, "Timeout should return Running initially")

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)
	result = timeout(ctx)
	assert.Equal(t, Failure, result, "Timeout should return Failure after duration")

	// Test with immediate success
	success := Action(func(ctx Context) Result { return Success })
	timeout = Timeout(1*time.Second, success)
	result = timeout(ctx)
	assert.Equal(t, Success, result, "Timeout should return Success when child succeeds quickly")
}

func TestDelay(t *testing.T) {
	ctx := make(Context)

	executed := false
	action := Action(func(ctx Context) Result {
		executed = true
		return Success
	})

	delay := Delay(3, action)

	// First tick - delay
	result := delay(ctx)
	assert.Equal(t, Running, result, "Delay should return Running on first tick")
	assert.False(t, executed, "Action should not be executed on first tick")

	// Second tick - delay
	result = delay(ctx)
	assert.Equal(t, Running, result, "Delay should return Running on second tick")
	assert.False(t, executed, "Action should not be executed on second tick")

	// Third tick - delay
	result = delay(ctx)
	assert.Equal(t, Running, result, "Delay should return Running on third tick")
	assert.False(t, executed, "Action should not be executed on third tick")

	// Fourth tick - execute
	result = delay(ctx)
	assert.Equal(t, Success, result, "Delay should return Success after delay")
	assert.True(t, executed, "Action should be executed after delay")
}

func TestLimiter(t *testing.T) {
	ctx := make(Context)

	callCount := 0
	action := Action(func(ctx Context) Result {
		callCount++
		return Success
	})

	limiter := Limiter(3, action)

	// First call - allowed
	result := limiter(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, 1, callCount)

	// Second call - allowed
	result = limiter(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, 2, callCount)

	// Third call - allowed
	result = limiter(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, 3, callCount)

	// Fourth call - blocked
	result = limiter(ctx)
	assert.Equal(t, Failure, result, "Limiter should block calls after maxCalls")
	assert.Equal(t, 3, callCount, "Call count should not increase after limit")
}
