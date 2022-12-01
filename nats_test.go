package connect

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/overmindtech/sdp-go"
)

func TestToNatsOptions(t *testing.T) {
	t.Run("with defaults", func(t *testing.T) {
		o := NATSOptions{}

		expectedOptions, err := optionsToStruct([]nats.Option{
			nats.Timeout(ConnectionTimeoutDefault),
			nats.MaxReconnects(MaxReconnectsDefault),
			nats.ReconnectWait(ReconnectWaitDefault),
			nats.ReconnectJitter(ReconnectJitterDefault, ReconnectJitterDefault),
			nats.DisconnectErrHandler(DisconnectErrHandlerDefault),
			nats.ReconnectHandler(ReconnectHandlerDefault),
			nats.ClosedHandler(ClosedHandlerDefault),
			nats.LameDuckModeHandler(LameDuckModeHandlerDefault),
			nats.ErrorHandler(ErrorHandlerDefault),
		})

		if err != nil {
			t.Fatal(err)
		}

		server, options := o.ToNatsOptions()

		if server != "" {
			t.Error("Expected server to be empty")
		}

		actualOptions, err := optionsToStruct(options)

		if err != nil {
			t.Fatal(err)
		}

		if expectedOptions.MaxReconnect != actualOptions.MaxReconnect {
			t.Errorf("Expected MaxReconnect to be %v, got %v", expectedOptions.MaxReconnect, actualOptions.MaxReconnect)
		}

		if expectedOptions.Timeout != actualOptions.Timeout {
			t.Errorf("Expected ConnectionTimeout to be %v, got %v", expectedOptions.Timeout, actualOptions.Timeout)
		}

		if expectedOptions.ReconnectWait != actualOptions.ReconnectWait {
			t.Errorf("Expected ReconnectWait to be %v, got %v", expectedOptions.ReconnectWait, actualOptions.ReconnectWait)
		}

		if expectedOptions.ReconnectJitter != actualOptions.ReconnectJitter {
			t.Errorf("Expected ReconnectJitter to be %v, got %v", expectedOptions.ReconnectJitter, actualOptions.ReconnectJitter)
		}

		// TokenClient
		if expectedOptions.UserJWT != nil || expectedOptions.SignatureCB != nil {
			t.Error("Expected TokenClient to be nil")
		}

		if actualOptions.DisconnectedErrCB == nil {
			t.Error("Expected DisconnectedErrCB to be non-nil")
		}

		if actualOptions.ReconnectedCB == nil {
			t.Error("Expected ReconnectedCB to be non-nil")
		}

		if actualOptions.ClosedCB == nil {
			t.Error("Expected ClosedCB to be non-nil")
		}

		if actualOptions.LameDuckModeHandler == nil {
			t.Error("Expected LameDuckModeHandler to be non-nil")
		}

		if actualOptions.AsyncErrorCB == nil {
			t.Error("Expected AsyncErrorCB to be non-nil")
		}
	})

	t.Run("with non-defaults", func(t *testing.T) {
		var disconnectErrHandlerUsed bool
		var reconnectHandlerUsed bool
		var closedHandlerUsed bool
		var lameDuckModeHandlerUsed bool
		var errorHandlerUsed bool

		o := NATSOptions{
			Servers:              []string{"one", "two"},
			ConnectionName:       "foo",
			MaxReconnects:        999,
			ReconnectWait:        999,
			ReconnectJitter:      999,
			DisconnectErrHandler: func(c *nats.Conn, err error) { disconnectErrHandlerUsed = true },
			ReconnectHandler:     func(c *nats.Conn) { reconnectHandlerUsed = true },
			ClosedHandler:        func(c *nats.Conn) { closedHandlerUsed = true },
			LameDuckModeHandler:  func(c *nats.Conn) { lameDuckModeHandlerUsed = true },
			ErrorHandler:         func(c *nats.Conn, s *nats.Subscription, err error) { errorHandlerUsed = true },
		}

		expectedOptions, err := optionsToStruct([]nats.Option{
			nats.Name("foo"),
			nats.MaxReconnects(999),
			nats.ReconnectWait(999),
			nats.ReconnectJitter(999, 999),
			nats.DisconnectErrHandler(nil),
			nats.ReconnectHandler(nil),
			nats.ClosedHandler(nil),
			nats.LameDuckModeHandler(nil),
			nats.ErrorHandler(nil),
		})

		if err != nil {
			t.Fatal(err)
		}

		server, options := o.ToNatsOptions()

		if server != "one,two" {
			t.Errorf("Expected server to be one,two got %v", server)
		}

		actualOptions, err := optionsToStruct(options)

		if err != nil {
			t.Fatal(err)
		}

		if expectedOptions.MaxReconnect != actualOptions.MaxReconnect {
			t.Errorf("Expected MaxReconnect to be %v, got %v", expectedOptions.MaxReconnect, actualOptions.MaxReconnect)
		}

		if expectedOptions.ReconnectWait != actualOptions.ReconnectWait {
			t.Errorf("Expected ReconnectWait to be %v, got %v", expectedOptions.ReconnectWait, actualOptions.ReconnectWait)
		}

		if expectedOptions.ReconnectJitter != actualOptions.ReconnectJitter {
			t.Errorf("Expected ReconnectJitter to be %v, got %v", expectedOptions.ReconnectJitter, actualOptions.ReconnectJitter)
		}

		if actualOptions.DisconnectedErrCB != nil {
			actualOptions.DisconnectedErrCB(nil, nil)
			if !disconnectErrHandlerUsed {
				t.Error("DisconnectErrHandler not used")
			}
		} else {
			t.Error("Expected DisconnectedErrCB to non-nil")
		}

		if actualOptions.ReconnectedCB != nil {
			actualOptions.ReconnectedCB(nil)
			if !reconnectHandlerUsed {
				t.Error("ReconnectHandler not used")
			}
		} else {
			t.Error("Expected ReconnectedCB to non-nil")
		}

		if actualOptions.ClosedCB != nil {
			actualOptions.ClosedCB(nil)
			if !closedHandlerUsed {
				t.Error("ClosedHandler not used")
			}
		} else {
			t.Error("Expected ClosedCB to non-nil")
		}

		if actualOptions.LameDuckModeHandler != nil {
			actualOptions.LameDuckModeHandler(nil)
			if !lameDuckModeHandlerUsed {
				t.Error("LameDuckModeHandler not used")
			}
		} else {
			t.Error("Expected LameDuckModeHandler to non-nil")
		}

		if actualOptions.AsyncErrorCB != nil {
			actualOptions.AsyncErrorCB(nil, nil, nil)
			if !errorHandlerUsed {
				t.Error("ErrorHandler not used")
			}
		} else {
			t.Error("Expected AsyncErrorCB to non-nil")
		}
	})
}

func TestNATSConnect(t *testing.T) {
	t.Run("with a bad URL", func(t *testing.T) {
		o := NATSOptions{
			Servers:    []string{"nats://badname.dontresolve.com"},
			NumRetries: 10,
			RetryDelay: 100 * time.Millisecond,
		}

		start := time.Now()

		_, err := o.Connect()

		if time.Since(start) > 2000*time.Millisecond {
			t.Errorf("Reconnecting took too long, expected <2s got: %v", time.Since(start).String())
		}

		switch err.(type) {
		case MaxRetriesError:
			// This is good
		default:
			t.Errorf("Unknown error type %T", err)
		}
	})

	t.Run("with a bad URL, but a good token", func(t *testing.T) {
		tk := GetTestOAuthTokenClient(t)

		startToken, err := tk.GetJWT()

		if err != nil {
			t.Fatal(err)
		}

		o := NATSOptions{
			Servers:     []string{"nats://badname.dontresolve.com"},
			TokenClient: tk,
			NumRetries:  3,
			RetryDelay:  100 * time.Millisecond,
		}

		_, err = o.Connect()

		switch err.(type) {
		case MaxRetriesError:
			// Make sure we have only got one token, not three
			currentToken, err := o.TokenClient.GetJWT()

			if err != nil {
				t.Fatal(err)
			}

			if currentToken != startToken {
				t.Error("Tokens have changed")
			}
		default:
			t.Errorf("Unknown error type %T", err)
		}
	})

	t.Run("with a good URL", func(t *testing.T) {
		o := NATSOptions{
			Servers: []string{
				"nats://nats:4222",
				"nats://localhost:4223",
			},
			NumRetries: 3,
			RetryDelay: 100 * time.Millisecond,
		}

		conn, err := o.Connect()

		if err != nil {
			t.Fatal(err)
		}

		ValidateNATSConnection(t, conn)
	})

	t.Run("with a good URL but no retries", func(t *testing.T) {
		o := NATSOptions{
			Servers: []string{
				"nats://nats:4222",
				"nats://localhost:4223",
			},
		}

		conn, err := o.Connect()

		if err != nil {
			t.Fatal(err)
		}

		ValidateNATSConnection(t, conn)
	})

	t.Run("with a good URL and infinite retries", func(t *testing.T) {
		o := NATSOptions{
			Servers: []string{
				"nats://nats:4222",
				"nats://localhost:4223",
			},
			NumRetries: -1,
			RetryDelay: 100 * time.Millisecond,
		}

		conn, err := o.Connect()

		if err != nil {
			t.Error(err)
		}

		ValidateNATSConnection(t, conn)
	})
}

func ValidateNATSConnection(t *testing.T, enc *nats.EncodedConn) {
	t.Helper()
	done := make(chan struct{})

	sub, err := enc.Subscribe("test", func(r *sdp.Response) {
		if r.Responder == "test" {
			done <- struct{}{}
		}
	})

	if err != nil {
		t.Error(err)
	}

	err = enc.Publish("test", &sdp.Response{
		Responder: "test",
		State:     sdp.ResponderState_COMPLETE,
	})

	if err != nil {
		t.Error(err)
	}

	// Wait for the message to come back
	select {
	case <-done:
		// Good
	case <-time.After(500 * time.Millisecond):
		t.Error("Didn't get message after 500ms")
	}

	err = sub.Unsubscribe()

	if err != nil {
		t.Error(err)
	}
}

func optionsToStruct(options []nats.Option) (nats.Options, error) {
	var o nats.Options
	var err error

	for _, option := range options {
		err = option(&o)

		if err != nil {
			return o, err
		}
	}

	return o, nil
}
