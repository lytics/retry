package retry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFailingRetry(t *testing.T) {
	n := 0
	// Always returning true to try again, should
	// eventually reach the max retries.
	X(4, 1*time.Millisecond, func() bool {
		n++
		return true
	})
	assert.Equal(t, 4, n)
}

func TestSuccessfulRetry(t *testing.T) {
	n := 0
	// Returing false from the function should
	// terminate the retries.
	X(4, 1*time.Millisecond, func() bool {
		n++
		if n == 2 {
			return false
		}
		return true
	})
	assert.Equal(t, 2, n)
}

func TestFailingRetryWithContext(t *testing.T) {
	n := 0
	ctx := context.Background()
	// Always returning true to try again, should
	// eventually reach the max retries.
	var ErrOops = errors.New("oops")
	err := XWithContext(ctx, 4, 1*time.Millisecond, func(ctx context.Context) error {
		n++
		return fmt.Errorf("failure %v, %w", n, ErrOops)
	})
	assert.Equal(t, 4, n)
	// Make sure error is expected
	assert.Equal(t, true, errors.Is(err, ErrOops))
	assert.Equal(t, true, strings.Contains(err.Error(), fmt.Sprintf("failure %v", n)))
}

func TestSuccessfulRetryWithContext(t *testing.T) {
	n := 0
	ctx := context.Background()
	// Returing false from the function should
	// terminate the retries.
	err := XWithContext(ctx, 4, 1*time.Millisecond, func(ctx context.Context) error {
		n++
		if n == 2 {
			return nil
		}
		return errors.New("oops")
	})
	assert.Equal(t, 2, n)
	assert.Nil(t, err)
}

func TestCancelledRetryWithContext(t *testing.T) {
	n := 0
	ctx, cancelFn := context.WithCancel(context.Background())
	// Always returning true to try again, should
	// eventually reach the max retries.
	err := XWithContext(ctx, 4, 1*time.Millisecond, func(ctx context.Context) error {
		n++
		if n == 2 {
			cancelFn()
		} else if n > 2 {
			time.Sleep(time.Minute)
		}
		return errors.New("oops")
	})
	assert.Equal(t, 2, n)
	// Make sure error is expected
	assert.Equal(t, true, errors.Is(err, context.Canceled))
}

func TestBackoff(t *testing.T) {
	const max = 8 * time.Second

	// A value of i less than 1 should be set to 1.
	// Large values of i should never return a
	// duration larger than max.
	for i := -10; i < 1000; i++ {
		assert.True(t, max >= backoff(i, max))
	}
}

func TestTailBackoff(t *testing.T) {
	const max = 8 * time.Second

	// Test that beyond the third try,
	// the max duration is returned.
	third := backoff(3, max)
	for i := 4; i < 1000; i++ {
		assert.Equal(t, third, backoff(i, max))
	}
}
