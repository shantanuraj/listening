package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const authEndpoint = "https://accounts.spotify.com/api/token"

var (
	clientID     = os.Getenv("SL_SPOTIFY_CLIENT_ID")
	clientSecret = os.Getenv("SL_SPOTIFY_CLIENT_SECRET")
)

func (c *Client) IsAuthenticated() bool {
	return c.token != nil && !c.token.HasExpired()
}

func (c *Client) Authenticate(ctx context.Context) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		authEndpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	str, _ := json.MarshalIndent(token, "", "  ")
	println(string(str))

	return &token, nil
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`

	CreatedAt time.Time `json:"-"` // Time when the token was created, not part of the JSON response
}

func (t TokenResponse) HasExpired() bool {
	now := time.Now()
	createdAt := t.CreatedAt
	return now.After(createdAt.Add(time.Second * time.Duration(t.ExpiresIn)))
}
