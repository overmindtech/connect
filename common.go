package multiconn

import "time"

type CommonOptions struct {
	NumRetries int
	RetryDelay time.Duration
}
