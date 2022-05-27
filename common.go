package multiconn

import "time"

type MaxRetriesError struct{}

func (m MaxRetriesError) Error() string {
	return "maximum retries reached"
}

type CommonOptions struct {
	NumRetries int           // How many times to retry connecting initially
	RetryDelay time.Duration // Delay between connection attempts
}
