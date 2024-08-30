package contracts

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
)

// HTTPOptions holds configuration options for HTTP requests.
type HTTPOptions struct {
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Port   string `json:"port,omitempty" yaml:"port,omitempty"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	Method string `json:"method,omitempty" yaml:"method,omitempty"`
	Schema string `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// WSOptions holds configuration options for WebSocket
type WSOptions struct {
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Port   string `json:"port,omitempty" yaml:"port,omitempty"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	Schema string `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// PingFunc defines a function type for executing a ping operation.
type PingFunc func(ctx context.Context) error

// Ping represents a ping operation with its related data and function.
type Ping struct {
	// Name of the ping operation
	Name string `json:"name" yaml:"name"`

	// Args - Arguments for the ping operation, can vary by implementation
	Args map[string]any `json:"args,omitempty" yaml:"args,omitempty"`

	// pingFunc - function to perform the ping action
	pingFunc PingFunc
}

// NewWSPinger creates a new Ping instance configured for WebSocket pinging.
func NewWSPinger(opts WSOptions) *Ping {
	return &Ping{
		pingFunc: DefaultWSPingFunc(opts),
	}
}

// NewHTTPPinger creates a new Ping instance configured for HTTP pinging.
func NewHTTPPinger(opts HTTPOptions) *Ping {

	return &Ping{
		pingFunc: DefaultHTTPPingFunc(opts),
	}
}

// Do executes the ping operation defined by the Ping instance
func (p *Ping) Do(ctx context.Context) error {
	return p.pingFunc(ctx)
}

// SetHandler allows setting a custom ping handler function.
func (p *Ping) SetHandler(handler func(ctx context.Context) error) {
	p.pingFunc = handler
}

// DefaultHTTPPingFunc returns a ping function for HTTP requests based on the provided options.
func DefaultHTTPPingFunc(opts HTTPOptions) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		host := opts.Host
		if opts.Port != "" {
			host = fmt.Sprintf("%s:%s", opts.Host, opts.Port)
		}

		// Construct the HTTP URL using the provided options
		httpURL := url.URL{
			Scheme: opts.Schema,
			Host:   host,
			Path:   opts.Path,
		}

		// Create a new HTTP request with the context
		req, err := http.NewRequestWithContext(ctx, opts.Method, httpURL.String(), nil)

		if err != nil {
			return err // Error handling for request creation
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err // Error handling for request execution
		}

		defer resp.Body.Close() // Ensure the response body is closed

		// Check if the response status code indicates success
		if resp.StatusCode != http.StatusOK {
			// Error handling for unsuccessful ping
			return fmt.Errorf("ping failed with status code %d", resp.StatusCode)
		}

		return nil // Return nil if the ping was successful
	}
}

// DefaultWSPingFunc returns a ping function for WebSocket based on provided options.
func DefaultWSPingFunc(opts WSOptions) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		host := opts.Host // Base host
		if opts.Port != "" {
			// Append port if it's specified
			host = fmt.Sprintf("%s:%s", opts.Host, opts.Port)
		}

		// Construct the WebSocket URL using the provided options
		wsURL := url.URL{
			Scheme: opts.Schema,
			Host:   host,
			Path:   opts.Path,
		}

		// Establish the WebSocket connection
		conn, _, err := websocket.DefaultDialer.Dial(
			wsURL.String(),
			nil,
		)
		if err != nil {
			return err // Error handling for connection failure
		}

		defer conn.Close() // Ensure the connection is closed

		return nil // Return nil if the ping was successful
	}
}
