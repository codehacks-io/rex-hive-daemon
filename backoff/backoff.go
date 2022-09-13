package backoff

import (
	"math"
	"time"
)

const backoffBaseDelaySeconds = 5
const BackoffResetIfUpSeconds = 600

func ExpBackoffSeconds(attempt int) time.Duration {
	// Cap to 5 minutes
	if attempt >= 6 {
		return time.Second * 300
	}

	if attempt < 0 {
		return 0
	}

	return time.Second * time.Duration(math.Pow(2, float64(attempt))*backoffBaseDelaySeconds)
}
