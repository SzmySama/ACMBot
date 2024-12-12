package helper

import "time"

/*
RateLimiter
*/
type RateLimiter []struct {
	rate time.Duration
}
