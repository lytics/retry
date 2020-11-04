// Package retry retries something X number of times, with an exponential backoff
// between each attempt. The backoff is calculated to reach the maximum backoff
// within three attempts.
//
// Example:
//     // Retry six times with a maximum backoff of 5 seconds
//     // between the retry attempts.
//
//     var err error
//     retry.X(6, 5*time.Second, func() bool {
//         err = DoSomething()
//         return err != nil
//     })
//     if err != nil {
//         // The error is not nil, so all retries failed.
//     } else {
//         // The error is nil, so one succeeded.
//     }
package retry

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// X number of retries. Function f should return false if it
// wants to stop trying, but never more than x calls of f
// are done. Calls to f have a sleep duration between them.
//
// Example 1:
//    var err error
//    retry.X(3, 5*time.Second, func() bool {
//        err = DoSomething()
//        return err != nil
//    })
//
// The use of "return err != nil" is an ideomatic way of
// returning true, keep trying, when the error is not nil.
func X(x int, maxBackoff time.Duration, f func() bool) {
	for i := 0; i < x; i++ {
		if i > 0 {
			time.Sleep(backoff(i, maxBackoff))
		}
		if !f() {
			return
		}
	}
}

// XWithContext runs function f until f returns nil or the
// number of retries exceeds x. Never more than x calls of f
// are done. Calls to f have a sleep duration between them.
// XWithError will return a wrapped error around f's errors
// if all attempts fail.
// The attempts can be cancelled with ctx. If f does not cancel
// when ctx is done, then the currently-running f will be allowed
// to complete first.
//
// Example 1:
//    var err error
//    retry.XWithError(3, 5*time.Second, func() error {
//        err = DoSomething()
//        return err
//    })
func XWithContext(ctx context.Context, x int, maxBackoff time.Duration, f func(ctx context.Context) error) error {
	var latestErr error
	for i := 0; i < x; i++ {
		timer := time.NewTimer(backoff(i, maxBackoff))
		defer timer.Stop()

		select {
		case <-ctx.Done():
			// context cancelled
			return ctx.Err()
		case <-timer.C:
			if latestErr = f(ctx); latestErr == nil {
				// finished ok!
				return nil
			}
		}
	}
	// ran out of retries
	return fmt.Errorf("all attempts failed, last attempt: %w", latestErr)
}

// backoff with exponential delay. On try 0, duration will be zero.
// Max will be reached in three tries. The min is a small but
// proportional fraction of the max, and a random jitter of
// between [0, min*try] is added when below max.
//
// Backoff is useful if you don't want to use the retry.X but want
// to calculate exponential backoff with jitter for your own use.
func backoff(try int, max time.Duration) time.Duration {
	if try < 1 {
		return time.Duration(0)
	}
	if try > 3 {
		return max
	}
	// 2^3 == 8. If you change this value then
	// you need to update the documentation.
	min := max / time.Duration(8)
	jit := int64(min) * int64(try)
	dur := time.Duration(min) << uint64(try)
	dur += time.Duration(rand.Int63n(jit))
	if dur < time.Duration(0) || dur > max {
		return max
	}
	return dur
}
