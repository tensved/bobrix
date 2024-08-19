package contracts

import (
	"context"
	"encoding/json"
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

type Ping struct {
	HTTPOpts *HTTPOptions `json:"http,omitempty" yaml:"http,omitempty"`
	WSOpts   *WSOptions   `json:"ws,omitempty" yaml:"ws,omitempty"`

	pingFuc func(ctx context.Context) error
}

func NewWSPinger(opts WSOptions) *Ping {
	return &Ping{
		pingFuc: DefaultWSPingFunc(opts),
	}
}

func NewHTTPPinger(opts HTTPOptions) *Ping {

	return &Ping{
		pingFuc: DefaultHTTPPingFunc(opts),
	}
}

func (p *Ping) Do(ctx context.Context) error {
	return p.pingFuc(ctx)
}

func (p *Ping) SetHandler(handler func(ctx context.Context) error) {
	p.pingFuc = handler
}

func (p *Ping) UnmarshalJSON(data []byte) error {
	var tempPing struct {
		HTTPOpts *HTTPOptions `json:"http,omitempty" yaml:"http,omitempty"`
		WSOpts   *WSOptions   `json:"ws,omitempty" yaml:"ws,omitempty"`
	}

	if err := json.Unmarshal(data, &tempPing); err != nil {
		return err
	}

	p.HTTPOpts = tempPing.HTTPOpts

	switch {
	case p.HTTPOpts != nil:
		p.pingFuc = func(ctx context.Context) error {
			req, err := http.NewRequestWithContext(ctx, p.HTTPOpts.Method,
				fmt.Sprintf("%s://%s:%s%s", p.HTTPOpts.Schema, p.HTTPOpts.Host, p.HTTPOpts.Port, p.HTTPOpts.Path),
				nil,
			)
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
	case p.WSOpts != nil:
		p.pingFuc = func(ctx context.Context) error {
			conn, _, err := websocket.DefaultDialer.Dial(
				fmt.Sprintf("%s://%s:%s%s", p.WSOpts.Schema, p.WSOpts.Host, p.WSOpts.Port, p.WSOpts.Path),
				nil,
			)
			if err != nil {
				return err
			}

			return conn.Close()
		}
	}

	return nil
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
