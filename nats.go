package connect

import (
	"strings"
	"time"

	"github.com/overmindtech/sdp-go/sdp"
	log "github.com/sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

// Defaults
const MaxReconnectsDefault = -1
const ReconnectWaitDefault = 1 * time.Second
const ReconnectJitterDefault = 5 * time.Second
const ConnectionTimeoutDefault = 10 * time.Second

type MaxRetriesError struct{}

func (m MaxRetriesError) Error() string {
	return "maximum retries reached"
}

var DisconnectErrHandlerDefault = func(c *nats.Conn, e error) {
	fields := log.Fields{
		"error": e,
	}

	if c != nil {
		fields["address"] = c.ConnectedAddr()
	}

	log.WithFields(fields).Error("NATS disconnected")
}
var ReconnectHandlerDefault = func(c *nats.Conn) {
	fields := log.Fields{}

	if c != nil {
		fields["reconnects"] = c.Reconnects
		fields["ServerID"] = c.ConnectedServerId()
		fields["URL"] = c.ConnectedUrl()

	}

	log.WithFields(fields).Info("NATS reconnected")
}
var ClosedHandlerDefault = func(c *nats.Conn) {
	fields := log.Fields{}

	if c != nil {
		fields["error"] = c.LastError()
	}

	log.WithFields(fields).Info("NATS connection closed")
}
var LameDuckModeHandlerDefault = func(c *nats.Conn) {
	fields := log.Fields{}

	if c != nil {
		fields["address"] = c.ConnectedAddr()

	}

	log.WithFields(fields).Info("NATS server has entered lame duck mode")
}
var ErrorHandlerDefault = func(c *nats.Conn, s *nats.Subscription, e error) {
	fields := log.Fields{
		"error": e,
	}

	if c != nil {
		fields["address"] = c.ConnectedAddr()
	}

	if s != nil {
		fields["subject"] = s.Subject
		fields["queue"] = s.Queue
	}

	log.WithFields(fields).Error("NATS error")
}

type NATSOptions struct {
	Servers              []string            // List of server to connect to
	ConnectionName       string              // The client name
	MaxReconnects        int                 // The maximum number of reconnect attempts
	ConnectionTimeout    time.Duration       // The timeout for Dial on a connection
	ReconnectWait        time.Duration       // Wait time between reconnect attempts
	ReconnectJitter      time.Duration       // The upper bound of a random delay added ReconnectWait
	TokenClient          TokenClient         // The client to use to get NATS tokens
	DisconnectErrHandler nats.ConnErrHandler // Runs when NATS is diconnected
	ReconnectHandler     nats.ConnHandler    // Runs when NATS has reconnected
	ClosedHandler        nats.ConnHandler    // Runs when a connection has been closed
	LameDuckModeHandler  nats.ConnHandler    // Runs when the connction enters "lame duck mode"
	ErrorHandler         nats.ErrHandler     // Runs when there is a NATS error
	AdditionalOptions    []nats.Option       // Addition options to pass to the connection
	NumRetries           int                 // How many times to retry connecting initially, use -1 to retry indefinitely
	RetryDelay           time.Duration       // Delay between connection attempts
}

// ToNatsOptions Converts the struct to connection string and a set of NATS
// options
func (o NATSOptions) ToNatsOptions() (string, []nats.Option) {
	serverString := strings.Join(o.Servers, ",")
	options := make([]nats.Option, 0)

	if o.ConnectionName != "" {
		options = append(options, nats.Name(o.ConnectionName))
	}

	if o.MaxReconnects != 0 {
		options = append(options, nats.MaxReconnects(o.MaxReconnects))
	} else {
		options = append(options, nats.MaxReconnects(MaxReconnectsDefault))
	}

	if o.ConnectionTimeout != 0 {
		options = append(options, nats.Timeout(o.ConnectionTimeout))
	} else {
		options = append(options, nats.Timeout(ConnectionTimeoutDefault))
	}

	if o.ReconnectWait != 0 {
		options = append(options, nats.ReconnectWait(o.ReconnectWait))
	} else {
		options = append(options, nats.ReconnectWait(ReconnectWaitDefault))
	}

	if o.ReconnectJitter != 0 {
		options = append(options, nats.ReconnectJitter(o.ReconnectJitter, o.ReconnectJitter))
	} else {
		options = append(options, nats.ReconnectJitter(ReconnectJitterDefault, ReconnectJitterDefault))
	}

	if o.TokenClient != nil {
		options = append(options, nats.UserJWT(o.TokenClient.GetJWT, o.TokenClient.Sign))
	}

	if o.DisconnectErrHandler != nil {
		options = append(options, nats.DisconnectErrHandler(o.DisconnectErrHandler))
	} else {
		options = append(options, nats.DisconnectErrHandler(DisconnectErrHandlerDefault))
	}

	if o.ReconnectHandler != nil {
		options = append(options, nats.ReconnectHandler(o.ReconnectHandler))
	} else {
		options = append(options, nats.ReconnectHandler(ReconnectHandlerDefault))
	}

	if o.ClosedHandler != nil {
		options = append(options, nats.ClosedHandler(o.ClosedHandler))
	} else {
		options = append(options, nats.ClosedHandler(ClosedHandlerDefault))
	}

	if o.LameDuckModeHandler != nil {
		options = append(options, nats.LameDuckModeHandler(o.LameDuckModeHandler))
	} else {
		options = append(options, nats.LameDuckModeHandler(LameDuckModeHandlerDefault))
	}

	if o.ErrorHandler != nil {
		options = append(options, nats.ErrorHandler(o.ErrorHandler))
	} else {
		options = append(options, nats.ErrorHandler(ErrorHandlerDefault))
	}

	options = append(options, o.AdditionalOptions...)

	return serverString, options
}

// Connect Connects to NATS using the supplied options, including retrying if
// unavailable
func (o NATSOptions) Connect() (sdp.EncodedConnection, error) {
	servers, opts := o.ToNatsOptions()

	var triesLeft int

	if o.NumRetries >= 0 {
		triesLeft = o.NumRetries + 1
	} else {
		triesLeft = -1
	}

	var nc *nats.Conn
	var err error

	for triesLeft != 0 {
		log.WithFields(log.Fields{
			"servers": servers,
		}).Info("NATS connecting")

		nc, err = nats.Connect(
			servers,
			opts...,
		)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Error connecting to NATS")

			triesLeft--
			time.Sleep(o.RetryDelay)
			continue
		}

		break
	}

	if err != nil {
		return &sdp.EncodedConnectionImpl{}, MaxRetriesError{}
	}

	return &sdp.EncodedConnectionImpl{Conn: nc}, nil
}
