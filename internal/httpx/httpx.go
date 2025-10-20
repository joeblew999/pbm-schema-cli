package httpx

import (
	"net/http"
	"time"
)

// Doer wraps http.Client for testability.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct { h Doer }

func New(timeout time.Duration) *Client {
	return &Client{h: &http.Client{Timeout: timeout}}
}

func WithDoer(d Doer) *Client { return &Client{h: d} }

func (c *Client) Do(req *http.Request) (*http.Response, error) { return c.h.Do(req) }
