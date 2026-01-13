package outbox

import (
	"math/rand"
	"time"
)

func backoff(retryCount int, baseDelay time.Duration) time.Duration {
	jitter := time.Duration(randInt(0, 1000)) * time.Millisecond
	b := time.Duration(1<<uint(retryCount))*baseDelay + jitter
	return b
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
