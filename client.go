package pvcusage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/monaco-io/request"
	"github.com/rs/zerolog/log"
)

// Client provides simplified access to the k8s API.
type Client struct {
	apiAddr   url.URL
	token     string
	tlsConfig *tls.Config
	timeout   time.Duration
}

// Option allows modification of the k8s API client.
type Option func(*Client)

// CA attempts to load a trusted Certificate Authority into the API client.
func CA(path string) Option {
	return func(c *Client) {
		info, err := os.Stat(path)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("failed to get path info")
		}
		if info.IsDir() {
			log.Info().
				Str("path", path).
				Msg("supplied path is a directory; assuming CA filename is ca.crt")
			path = filepath.Join(path, "ca.crt")
		}

		caBytes, err := os.ReadFile(path)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("failed to read path")
			c.tlsConfig.InsecureSkipVerify = true
			return
		}

		log.Info().
			Str("path", path).
			Msg("creating CA pool")
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caBytes); !ok {
			c.tlsConfig.InsecureSkipVerify = true
		} else {
			c.tlsConfig.RootCAs = pool
		}
	}
}

// Timeout customizes the maximum amount of time to wait for a response from the k8s API.
func Timeout(dur time.Duration) Option {
	return func(c *Client) {
		c.timeout = dur
	}
}

// New creates a new k8s API client.
func New(addr, token string) *Client {
	return &Client{
		apiAddr: url.URL{
			Scheme: "https",
			Host:   addr,
			Path:   "/api/v1",
		},
		token:     token,
		tlsConfig: new(tls.Config),
		timeout:   5 * time.Second,
	}
}

// With enables modification of the k8s API client configuration.
func (c *Client) With(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

// Req prepares a request against the k8s API.
func (c *Client) Req(ctx context.Context, parts ...string) *request.Client {
	return &request.Client{
		Context:   ctx,
		URL:       c.apiAddr.JoinPath(parts...).String(),
		TLSConfig: c.tlsConfig,
		Method:    "GET",
		Bearer:    c.token,
		Timeout:   c.timeout,
	}
}
