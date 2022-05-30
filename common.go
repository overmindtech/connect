package multiconn

import (
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dgraph-io/dgo/v210"
	"github.com/nats-io/nats.go"
)

const ConnectionTimeoutDefault = 10 * time.Second

type MaxRetriesError struct{}

func (m MaxRetriesError) Error() string {
	return "maximum retries reached"
}

type CommonOptions struct {
	NumRetries int           // How many times to retry connecting initially
	RetryDelay time.Duration // Delay between connection attempts
}

// TotalTries Returns the total number of times we shoyuld try to connect,
// including the first try. For positive numbers this is retries + 1, for
// negative we just leave it alone
func (c CommonOptions) TotalTries() int {
	if c.NumRetries >= 0 {
		return c.NumRetries + 1
	} else {
		return c.NumRetries
	}

}

// ConnectionOptions Options for connecting to each service, if these are nil,
// the service won't be connected
type ConnectionOptions struct {
	NATS   *NATSConnectionOptions
	DGraph *DGraphConnectionOptions
}

type Connections struct {
	NATS   *nats.EncodedConn
	DGraph *dgo.Dgraph
}

// Connect Connects to the services for which options were supplied
func (c *ConnectionOptions) Connect() (Connections, error) {
	var wg sync.WaitGroup
	var enc *nats.EncodedConn
	var d *dgo.Dgraph

	if c.NATS != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			enc, err = c.NATS.Connect()

			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("NATS connection failed, giving up")
			}
		}()
	}

	if c.DGraph != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			d, err = c.DGraph.Connect()

			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("DGraph connection failed, giving up")
			}
		}()
	}

	wg.Wait()

	var err error

	if enc == nil && d == nil {
		err = errors.New("all connections failed")
	}

	return Connections{
		NATS:   enc,
		DGraph: d,
	}, err
}
