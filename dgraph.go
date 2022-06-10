package multiconn

import (
	"context"
	"time"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"

	log "github.com/sirupsen/logrus"
)

type DGraphConnectionOptions struct {
	CommonOptions

	Servers           []string // List of Alphas servers to try connecting to, if multiple work, the connection will be created to multiple alphas. Format is `host:port`
	ConnectionTimeout time.Duration
}

func (d DGraphConnectionOptions) Connect() (*dgo.Dgraph, error) {
	var dGraphClient *dgo.Dgraph
	tries := d.TotalTries()

	if d.ConnectionTimeout == 0 {
		d.ConnectionTimeout = ConnectionTimeoutDefault
	}

	for tries != 0 {
		connections := make([]*grpc.ClientConn, 0)

		for _, server := range d.Servers {
			dialOpts := append([]grpc.DialOption{},
				grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
				grpc.WithBlock(),
				grpc.FailOnNonTempDialError(true),
			)

			// TODO: Once we support auth for DGraph we'll need to fill in this
			// conditional depending on whether we're using auth or not.
			// Presumably we'll also need certificate config and things like
			// that
			if true {
				dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
			}

			log.WithFields(log.Fields{
				"server": server,
			}).Info("Connecting to DGraph")

			ctx, cancel := context.WithTimeout(context.Background(), d.ConnectionTimeout)

			defer cancel()

			c, err := grpc.DialContext(ctx, server, dialOpts...)

			if err != nil {
				log.WithFields(log.Fields{
					"server": server,
					"error":  err.Error(),
				}).Error("DGraph connection failed")
			} else {
				log.WithFields(log.Fields{
					"server": server,
				}).Info("DGraph connecton opened")

				connections = append(connections, c)
			}
		}

		if len(connections) == 0 {
			log.WithFields(log.Fields{
				"servers": d.Servers,
			}).Error("Could not connect to any DGraph endpoints")

			tries--
			time.Sleep(d.RetryDelay)
			continue
		}

		apiClients := make([]api.DgraphClient, len(connections))
		serverFields := make(log.Fields)

		for i, conn := range connections {
			var versionString string

			apiClients[i] = api.NewDgraphClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Check the version and add to log fields
			version, err := apiClients[i].CheckVersion(ctx, &api.Check{})

			if err != nil || version == nil {
				versionString = "unknown version"
			} else {
				versionString = version.Tag
			}

			serverFields[conn.Target()] = versionString
		}

		dGraphClient = dgo.NewDgraphClient(apiClients...)

		log.WithFields(serverFields).Info("DGraph client created")

		break
	}

	if dGraphClient == nil {
		return nil, MaxRetriesError{}
	}

	return dGraphClient, nil
}
