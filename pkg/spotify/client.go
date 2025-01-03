package spotify

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Client struct {
	token      *TokenResponse
	httpClient *http.Client
}

const host = "https://api.spotify.com/v1"

var DefaultClient = &Client{
	httpClient: &http.Client{
		Timeout: time.Second * 10,
	},
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, "GET", path, nil)
}

func (c *Client) Put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, "PUT", path, body)
}

func (c *Client) do(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
) (*http.Response, error) {
	url := host + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": []string{"Bearer " + c.token.AccessToken},
	}

	return c.httpClient.Do(req)
}

func (c *Client) SetToken(token *TokenResponse) {
	c.token = token
}
