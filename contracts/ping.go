package contracts

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
)

type HTTPOptions struct {
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Port   string `json:"port,omitempty" yaml:"port,omitempty"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	Method string `json:"method,omitempty" yaml:"method,omitempty"`
	Schema string `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type WSOptions struct {
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Port   string `json:"port,omitempty" yaml:"port,omitempty"`
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	Schema string `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type PingFunc func(ctx context.Context) error

type Ping struct {
	Name string         `json:"name" yaml:"name"`
	Args map[string]any `json:"args,omitempty" yaml:"args,omitempty"`

	pingFunc PingFunc
}

func NewWSPinger(opts WSOptions) *Ping {
	return &Ping{
		pingFunc: DefaultWSPingFunc(opts),
	}
}

func NewHTTPPinger(opts HTTPOptions) *Ping {

	return &Ping{
		pingFunc: DefaultHTTPPingFunc(opts),
	}
}

func (p *Ping) Do(ctx context.Context) error {
	return p.pingFunc(ctx)
}

func (p *Ping) SetHandler(handler func(ctx context.Context) error) {
	p.pingFunc = handler
}

func DefaultHTTPPingFunc(opts HTTPOptions) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		host := opts.Host
		if opts.Port != "" {
			host = fmt.Sprintf("%s:%s", opts.Host, opts.Port)
		}

		httpURL := url.URL{
			Scheme: opts.Schema,
			Host:   host,
			Path:   opts.Path,
		}

		req, err := http.NewRequestWithContext(ctx, opts.Method, httpURL.String(), nil)

		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("ping failed with status code %d", resp.StatusCode)
		}

		return nil
	}
}

func DefaultWSPingFunc(opts WSOptions) func(ctx context.Context) error {
	return func(ctx context.Context) error {

		host := opts.Host
		if opts.Port != "" {
			host = fmt.Sprintf("%s:%s", opts.Host, opts.Port)
		}

		wsURL := url.URL{
			Scheme: opts.Schema,
			Host:   host,
			Path:   opts.Path,
		}

		conn, _, err := websocket.DefaultDialer.Dial(
			wsURL.String(),
			nil,
		)
		if err != nil {
			return err
		}

		return conn.Close()
	}
}
