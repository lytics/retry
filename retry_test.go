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

func TestXFailure(t *testing.T) {
	t.Parallel()
	n := 0
	// Always returning true to try again, should
	// eventually reach the max retries.
	X(4, time.Millisecond, func() bool {
		n++
		return true
	})
	assert.Equal(t, 5, n)
}

func TestXSuccess(t *testing.T) {
	t.Parallel()
	n := 0
	// Returing false from the function should
	// terminate the retries.
	X(4, time.Millisecond, func() bool {
		n++
		return n != 2
	})
	assert.Equal(t, 2, n)
}

func TestXWithContextFailure(t *testing.T) {
	t.Parallel()
	n := 0
	ctx := context.Background()
	// Always returning true to try again, should
	// eventually reach the max retries.
	var ErrOops = errors.New("oops")
	err := XWithContext(ctx, 4, time.Millisecond, func(context.Context) error {
		n++
		return fmt.Errorf("failure %v, %w", n, ErrOops)
	})
	assert.Equal(t, 5, n)
	// Make sure error is expected
	assert.True(t, errors.Is(err, ErrOops))
	assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("failure %v", n)))
}

func TestXWithContextSuccess(t *testing.T) {
	n := 0
	ctx := context.Background()
	// Returing false from the function should
	// terminate the retries.
	err := XWithContext(ctx, 4, time.Millisecond, func(context.Context) error {
		n++
		if n == 2 {
			return nil
		}
		return errors.New("oops")
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestXWithContextCancelled(t *testing.T) {
	n := 0
	ctx, cancelFn := context.WithCancel(context.Background())
	// Always returning true to try again, should
	// eventually reach the max retries.
	err := XWithContext(ctx, 4, time.Millisecond, func(context.Context) error {
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
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestRetryWithContextNoRetries(t *testing.T) {
	t.Parallel()

	// Given
	var n int
	someErr := errors.New("oops")

	// When
	err := XWithContext(context.Background(), 0, 0, func(context.Context) error {
		n++
		return someErr
	})

	// Then
	assert.True(t, errors.Is(err, someErr))
	assert.Equal(t, 1, n)
}

func TestRetryWithContextBadN(t *testing.T) {
	t.Parallel()

	// Given
	var n int

	// When
	err := XWithContext(context.Background(), -1, 0, func(context.Context) error {
		n++
		return nil
	})

	// Then
	assert.Error(t, err)
	assert.Zero(t, n)
}

func TestRetryWithContextBadMaxBackoff(t *testing.T) {
	t.Parallel()

	// Given
	var n int

	// When
	err := XWithContext(context.Background(), 0, -1, func(context.Context) error {
		n++
		return nil
	})

	// Then
	assert.Error(t, err)
	assert.Zero(t, n)
}

func TestBackoff(t *testing.T) {
	t.Parallel()
	const max = 8 * time.Second

	// A value of i less than 1 should be set to 1.
	// Large values of i should never return a
	// duration larger than max.
	for i := -10; i < 1000; i++ {
		assert.True(t, max >= backoff(i, max))
	}
}

func TestBackoffTail(t *testing.T) {
	t.Parallel()
	const max = 8 * time.Second

	// Test that beyond the third try,
	// the max duration is returned.
	third := backoff(3, max)
	for i := 4; i < 1000; i++ {
		assert.Equal(t, third, backoff(i, max))
	}
}
