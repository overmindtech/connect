# Connect

`connect` manages the NATS connections used by Overmind infrastructure to the NATS network. This includes:

* Connecting
* Reconnects
* Error handing
* Auth

## Auth

In order to connect to NATS, a user needs an [Nkey](https://github.com/nats-io/nkeys) and a JWT. This JWT is issued by the [Overmind API](https://overmindtech.github.io/api-server/#tag/core/operation/CreateToken) based on the permissions that the user calling the API has in the OAuth token they supply when calling it. This OAuth token comes from one of the following flows:

### Client Credentials Flow

If the thing that is connecting already has the required permissions, then you can use the [client credentials flow](https://auth0.com/docs/get-started/authentication-and-authorization-flow/client-credentials-flow) to get a token. If the application whose Client ID you supply is linked to a specific Organization then you can let the API connect you automatically:

```go
flowConfig := ClientCredentialsConfig{
    ClientID:     "SOMETHING",
    ClientSecret: "SECRET",
}

client := NewOAuthTokenClient(
    fmt.Sprintf("https://%v/oauth/token", domain),
    exchangeURL,
    flowConfig,
)

o := NATSOptions{
    Servers:     []string{"nats://something"},
    TokenClient: client,
    NumRetries:  3,
    RetryDelay:  100 * time.Millisecond,
}

conn, err := o.Connect()
```

If however you need to connect to a specific Org, then you'll need an application with `admin:write` permissions, and to supply the org you want to connect to:

```go
flowConfig := ClientCredentialsConfig{
    ClientID:     "SOMETHING",
    ClientSecret: "SECRET",
    Org:          "org_somethingHere",
}

client := NewOAuthTokenClient(
    fmt.Sprintf("https://%v/oauth/token", domain),
    exchangeURL,
    flowConfig,
)

o := NATSOptions{
    Servers:     []string{"nats://something"},
    TokenClient: client,
    NumRetries:  3,
    RetryDelay:  100 * time.Millisecond,
}

conn, err := o.Connect()
```
