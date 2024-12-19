package helper

import "time"

type RateLimiter []struct {
	rate time.Duration
}
