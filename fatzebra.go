package fatzebra

import (
	"errors"
	"net/http"
)

// Endpoints provided for convenience.
const (
	SandboxEndpoint    = "gateway.pmnts-sandbox.io"
	ProductionEndpoint = "gateway.pmnts.io"
)

// Possible errors returned.
var (
	ErrBadHash  = errors.New("fatzebra: bad hash")
	ErrNotFound = errors.New("fatzebra: not found")
)

// Client represents a client.
type Client struct {
	client    *http.Client
	username  string
	password  string
	host      string
	secret    []byte
	maxAmount AUD
}

// ClientOpts represents the options usd to construct a Fat Zebra client.
type ClientOpts struct {
	Username  string
	Password  string
	Secret    string
	Host      string
	MaxAmount AUD // Provides last chance protection against charging
	// unintentionally large amounts in a single transaction.
	Client *http.Client
}

// NewClient returns a new client with the given options.
func NewClient(opts *ClientOpts) *Client {
	if opts.MaxAmount <= 0 {
		panic("fatzebra: maximum amount must be greater than 0")
	}

	c := &Client{
		username:  opts.Username,
		password:  opts.Password,
		host:      opts.Host,
		secret:    []byte(opts.Secret),
		maxAmount: opts.MaxAmount,
	}

	httpC := *(opts.Client)
	httpC.Transport = &Transport{
		c:    c,
		base: httpC.Transport,
	}

	c.client = &httpC

	return c
}
