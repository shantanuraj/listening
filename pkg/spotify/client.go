package spotify

import (
	"context"
	"net/http"
	"os"
	"time"
)

type Client struct {
	authToken  string
	userAgent  string
	httpClient *http.Client
}

const host = "https://api.spotify.com/v1"

var DefaultClient = &Client{
	authToken: "Bearer " + os.Getenv("SL_TOKEN"),
	userAgent: "sraj.me/listening",
	httpClient: &http.Client{
		Timeout: time.Second * 10,
	},
}

func (c Client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", host+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": []string{c.authToken},
		"User-Agent":    []string{c.userAgent},
	}

	return c.httpClient.Do(req)
}
