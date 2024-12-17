package spotify

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/shantanuraj/listening/pkg/dirs"
	"github.com/shantanuraj/listening/pkg/log"
)

const (
	spotifyAuthURL  = "https://accounts.spotify.com/authorize"
	spotifyTokenURL = "https://accounts.spotify.com/api/token"
	scope           = "user-read-currently-playing user-read-playback-state user-modify-playback-state user-read-recently-played"
)

var (
	clientID     = os.Getenv("SL_SPOTIFY_CLIENT_ID")
	clientSecret = os.Getenv("SL_SPOTIFY_CLIENT_SECRET")
)

func redirectURL(addr string) string {
	return fmt.Sprintf("%s%s", addr, "/callback")
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func authURL(addr string, state string) string {
	return fmt.Sprintf(
		"%s?response_type=code&client_id=%s&redirect_uri=%s&state=%s&scope=%s",
		spotifyAuthURL,
		clientID,
		redirectURL(addr),
		state,
		scope,
	)
}

func (c *Client) IsAuthenticated() bool {
	return c.token != nil && !c.token.HasExpired()
}

func (c *Client) IsTokenExpired() bool {
	if c.token == nil {
		return false
	}
	return c.token.HasExpired()
}

func (c *Client) RegisterAuthenticationHandlers(
	addr string,
	mux *http.ServeMux,
) error {
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("missing client ID or client secret")
	}

	credentialsPath, err := dirs.CredentialsPath()
	if err != nil {
		return fmt.Errorf("failed to get credentials path: %w", err)
	}

	if !c.IsAuthenticated() {
		token, err := loadToken(credentialsPath)
		if err != nil {
			return fmt.Errorf("failed to load persisted token: %w", err)
		}
		if token != nil {
			c.token = token
			log.Infof("Authenticated as %s", token.AccessToken[:8])
		}
	}

	state := generateState()

	mux.Handle(
		"GET /",
		http.RedirectHandler(authURL(addr, state), http.StatusTemporaryRedirect),
	)
	mux.Handle("GET /callback", spotifyCallbackHandler(c, state, addr, credentialsPath))
	mux.Handle("POST /refresh", refreshHandler(c))

	return nil
}

func spotifyCallbackHandler(
	c *Client,
	state string,
	addr string,
	credentialsPath string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		code := query.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		token, err := c.ExchangeCodeForToken(ctx, addr, code)
		if err != nil {
			http.Error(w, "failed to exchange code for token", http.StatusInternalServerError)
			return
		}

		c.token = token
		log.Infof("Authenticated as %s", token.AccessToken[:8])

		if err := saveToken(token, credentialsPath); err != nil {
			log.Errorf("failed to save token: %v", err)
		}

		http.Redirect(w, r, "/current", http.StatusTemporaryRedirect)
	}
}

func saveToken(token *TokenResponse, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(token)
}

func loadToken(path string) (*TokenResponse, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token TokenResponse
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (c *Client) RefreshToken(ctx context.Context) error {
	if c.token == nil {
		return fmt.Errorf("no token to refresh")
	}

	refreshToken := c.token.RefreshToken

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		spotifyTokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh: invalid status code: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return err
	}
	if token.AccessToken == "" {
		log.Errorf("invalid token response: %v", token)
		return fmt.Errorf("invalid token response")
	}
	token.CreatedAt = time.Now()
	if token.RefreshToken == "" && refreshToken != "" {
		token.RefreshToken = refreshToken
	}

	c.token = &token
	log.Infof("Authenticated as %s", token.AccessToken[:8])

	credentialsPath, err := dirs.CredentialsPath()
	if err != nil {
		return fmt.Errorf("failed to get credentials path: %w", err)
	}

	if err := saveToken(&token, credentialsPath); err != nil {
		log.Errorf("failed to save token: %v", err)
	}

	return nil
}

func (c *Client) ExchangeCodeForToken(
	ctx context.Context,
	addr string,
	code string,
) (*TokenResponse, error) {
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing client ID or client secret")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL(addr))
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		spotifyTokenURL,
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange: invalid status code: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	token.CreatedAt = time.Now()

	return &token, nil
}

func refreshHandler(c *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := c.RefreshToken(ctx); err != nil {
			log.Errorf("failed to refresh token: %v", err)
			http.Error(w, "failed to refresh token", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`

	CreatedAt time.Time `json:"created_at"` // Time when the token was created, not part of the JSON response
}

func (t TokenResponse) HasExpired() bool {
	now := time.Now()
	createdAt := t.CreatedAt
	return now.After(createdAt.Add(time.Second * time.Duration(t.ExpiresIn)))
}
