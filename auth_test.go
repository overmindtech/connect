package multiconn

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nkeys"
	"github.com/overmindtech/tokenx-client"
)

var tokenExchangeURLs = []string{
	"http://nats-token-exchange:8080/api/nats",
	"http://localhost:8080/api/nats",
}

func TestBasicTokenClient(t *testing.T) {
	keys, err := nkeys.CreateUser()

	if err != nil {
		t.Fatal(err)
	}

	c := NewBasicTokenClient("tokeny_mc_tokenface", keys)

	var token string

	token, err = c.GetJWT()

	if err != nil {
		t.Error(err)
	}

	if token != "tokeny_mc_tokenface" {
		t.Error("token mismatch")
	}

	data := []byte{1, 156, 230, 4, 23, 175, 11}

	signed, err := c.Sign(data)

	if err != nil {
		t.Fatal(err)
	}

	err = keys.Verify(data, signed)

	if err != nil {
		t.Error(err)
	}
}

func GetTestOAuthTokenClient(t *testing.T) *OAuthTokenClient {
	var domain string
	var clientID string
	var clientSecret string
	var exists bool

	errorFormat := "environment variable %v not found. Set up your test environment first. See: https://github.com/overmindtech/auth0-test-data"

	// Read secrets form the environment
	if domain, exists = os.LookupEnv("OVERMIND_NTE_ALLPERMS_DOMAIN"); !exists || domain == "" {
		t.Errorf(errorFormat, "OVERMIND_NTE_ALLPERMS_DOMAIN")
		t.Skip("Skipping due to missing environment setup")
	}

	if clientID, exists = os.LookupEnv("OVERMIND_NTE_ALLPERMS_CLIENT_ID"); !exists || clientID == "" {
		t.Errorf(errorFormat, "OVERMIND_NTE_ALLPERMS_CLIENT_ID")
		t.Skip("Skipping due to missing environment setup")
	}

	if clientSecret, exists = os.LookupEnv("OVERMIND_NTE_ALLPERMS_CLIENT_SECRET"); !exists || clientSecret == "" {
		t.Errorf(errorFormat, "OVERMIND_NTE_ALLPERMS_CLIENT_SECRET")
		t.Skip("Skipping due to missing environment setup")
	}

	exchangeURL, err := GetWorkingTokenExchange()

	if err != nil {
		t.Fatal(err)
	}

	return NewOAuthTokenClient(
		clientID,
		clientSecret,
		fmt.Sprintf("https://%v/oauth/token", domain),
		exchangeURL,
	)
}

func TestOAuthTokenClient(t *testing.T) {
	c := GetTestOAuthTokenClient(t)

	EnsureTestAccount(c.natsClient.AuthApi)

	var err error

	_, err = c.GetJWT()

	if err != nil {
		t.Error(err)
	}

	// Make sure it can sign
	data := []byte{1, 156, 230, 4, 23, 175, 11}

	_, err = c.Sign(data)

	if err != nil {
		t.Fatal(err)
	}

}

var TestAccountCreated bool

func EnsureTestAccount(a *tokenx.AuthApiService) error {
	if !TestAccountCreated {
		// This is the account that OAuth embeds in test tokens and therefore must
		// be created
		name := "test-account"

		req := a.AccountsPost(context.Background()).AccountRequestData(tokenx.AccountRequestData{
			Name: &name,
		})

		_, _, err := req.Execute()

		if err != nil {
			return err
		}

		TestAccountCreated = true
	}

	return nil
}

func GetWorkingTokenExchange() (string, error) {
	var err error

	for _, url := range tokenExchangeURLs {
		if err = testURL(url); err == nil {
			return url, nil
		}
	}

	return "", fmt.Errorf("no working token exchanges found: %v", err)
}

func testURL(testURL string) error {
	url, err := url.Parse(testURL)

	if err != nil {
		return fmt.Errorf("could not parse NATS URL: %v. Error: %v", testURL, err)
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(url.Hostname(), url.Port()), time.Second)

	if err == nil {
		conn.Close()
		return nil
	}

	return err
}
