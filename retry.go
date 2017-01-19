package retry

import (
	"math/rand"
	"time"
)

// X number of retries. Function f should return false if it
// wants to stop trying, but never more than x calls of f
// are done. Calls to f have a sleep duration between them.
// The Backoff function is used to calculate sleep time.
//
// Example 1:
//    var err error
//    var off uint64
//    retry.X(3, 5*time.Second, func() bool {
//        off, err = events.LastOffset()
//        return err != nil
//    })
//    ... do something with err ...
//
// The use of "return err != nil" is an ideomatic way of
// returning true, keep trying, when the error is not nil.
func X(x int, maxBackoff time.Duration, f func() bool) {
	for i := 0; i < x; i++ {
		if i > 0 {
			time.Sleep(Backoff(i, maxBackoff))
		}
		if !f() {
			return
		}
	}
}

// Backoff with exponential delay. On try 0, duration will be zero.
// Max will be reached in three tries. The min is a small but
// proportional fraction of the max, and a random jitter of
// between [0, min*try] is added when below max.
//
// Backoff is useful if you don't want to use the retry.X but want
// to calculate exponential backoff with jitter for your own use.
func Backoff(try int, max time.Duration) time.Duration {
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
