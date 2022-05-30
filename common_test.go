package multiconn

import (
	"testing"
	"time"
)

func TestMultiConnect(t *testing.T) {
	t.Run("with good connections", func(t *testing.T) {
		o := ConnectionOptions{
			NATS: &NATSConnectionOptions{
				Servers: []string{
					"nats://nats:4222",
					"nats://localhost:4223",
				},
				CommonOptions: CommonOptions{
					NumRetries: 3,
					RetryDelay: 100 * time.Millisecond,
				},
			},
			DGraph: &DGraphConnectionOptions{
				Servers: []string{
					"dgraph:9080",
					"localhost:9080",
				},
				CommonOptions: CommonOptions{
					NumRetries: 3,
					RetryDelay: 100 * time.Millisecond,
				},
			},
		}

		conns, err := o.Connect()

		if err != nil {
			t.Error(err)
		}

		if conns.DGraph == nil {
			t.Error("DGraph connection is nil")
		} else {
			ValidateDGraphConnection(t, conns.DGraph)
		}

		if conns.NATS == nil {
			t.Error("NATS connection is nil")
		} else {
			ValidateNATSConnection(t, conns.NATS)
		}
	})

	t.Run("with bad connections", func(t *testing.T) {
		o := ConnectionOptions{
			NATS: &NATSConnectionOptions{
				Servers: []string{
					"nats://foo:4222",
				},
				CommonOptions: CommonOptions{
					NumRetries: 3,
					RetryDelay: 100 * time.Millisecond,
				},
			},
			DGraph: &DGraphConnectionOptions{
				Servers: []string{
					"foo:9080",
				},
				CommonOptions: CommonOptions{
					NumRetries: 3,
					RetryDelay: 100 * time.Millisecond,
				},
			},
		}

		conns, err := o.Connect()

		if err == nil {
			t.Error("expected err got nil")
		}

		if conns.DGraph != nil {
			t.Error("DGraph connection is non-nil")
		}

		if conns.NATS != nil {
			t.Error("NATS connection is non-nil")
		}
	})

}
