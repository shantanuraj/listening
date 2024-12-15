package spotify

import (
	"context"
	"fmt"
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
	url := host + path
	fmt.Println(url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
