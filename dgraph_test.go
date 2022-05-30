package multiconn

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

func TestDGraphConnect(t *testing.T) {
	t.Run("with a bad URL", func(t *testing.T) {
		o := DGraphConnectionOptions{
			Servers: []string{"badname.dontresolve.com:443"},
			CommonOptions: CommonOptions{
				NumRetries: 10,
				RetryDelay: 100 * time.Millisecond,
			},
		}

		start := time.Now()

		_, err := o.Connect()

		if time.Since(start) > 1500*time.Millisecond {
			t.Errorf("Reconnecting took too long, expected >1.5s got: %v", time.Since(start).String())
		}

		switch err.(type) {
		case MaxRetriesError:
			// This is good
		default:
			t.Errorf("Unknown error type %T", err)
		}
	})

	t.Run("with a good URL", func(t *testing.T) {
		o := DGraphConnectionOptions{
			Servers: []string{
				"dgraph:9080",
				"localhost:9080",
			},
			CommonOptions: CommonOptions{
				NumRetries: 3,
				RetryDelay: 100 * time.Millisecond,
			},
		}

		client, err := o.Connect()

		if err != nil {
			t.Error(err)
		}

		ValidateDGraphConnection(t, client)
	})

	t.Run("with a good URL and defaults", func(t *testing.T) {
		o := DGraphConnectionOptions{
			Servers: []string{
				"dgraph:9080",
				"localhost:9080",
			},
		}

		client, err := o.Connect()

		if err != nil {
			t.Error(err)
		}

		ValidateDGraphConnection(t, client)
	})

	t.Run("with a good URL and infinite retrues", func(t *testing.T) {
		o := DGraphConnectionOptions{
			Servers: []string{
				"dgraph:9080",
				"localhost:9080",
			},
			CommonOptions: CommonOptions{
				NumRetries: -1,
				RetryDelay: 100 * time.Millisecond,
			},
		}

		client, err := o.Connect()

		if err != nil {
			t.Error(err)
		}

		ValidateDGraphConnection(t, client)
	})
}

func ValidateDGraphConnection(t *testing.T, c *dgo.Dgraph) {
	t.Helper()

	// Run a test query to make sure it worked
	query := `{
			test(func: uid(0x394c)) {
				uid
			}
		}`

	_, err := c.NewTxn().Do(context.Background(), &api.Request{
		Query: query,
	})

	if err != nil {
		t.Error(err)
	}
}
