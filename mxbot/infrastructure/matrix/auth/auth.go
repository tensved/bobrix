package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type synapseCreateUserReq struct {
	Password    string `json:"password"`
	Admin       bool   `json:"admin"`
	Deactivated bool   `json:"deactivated,omitempty"`
	Displayname string `json:"displayname,omitempty"`
	Threepids   any    `json:"threepids,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

func (a *Service) registerBot(ctx context.Context) error {
	if a.creds.AdminToken == "" {
		return fmt.Errorf("admin token is empty: cannot register user via synapse admin api")
	}

	// Collect the full MXID. If the Username already contains "@", we assume it's an MXID.
	mxid := a.creds.Username
	if !strings.HasPrefix(mxid, "@") {
		hs, err := url.Parse(a.client.HomeserverURL.String())
		if err != nil {
			return fmt.Errorf("parse homeserver url: %w", err)
		}
		mxid = fmt.Sprintf("@%s:%s", a.creds.Username, hs.Hostname())
	}

	endpoint := a.client.HomeserverURL.String() +
		"/_synapse/admin/v2/users/" + url.PathEscape(mxid)

	body, _ := json.Marshal(synapseCreateUserReq{
		Password:    a.creds.Password,
		Admin:       false,
		Displayname: a.name,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.creds.AdminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("admin api request failed: %w", err)
	}
	defer resp.Body.Close()

	// 201 Created — the user has been created
	// 200 OK — the user already exists (Synapse can also respond this way)
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return nil
	}

	return fmt.Errorf("synapse admin create user failed: status=%s", resp.Status)
}
