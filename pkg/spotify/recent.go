package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const recentEndpoint = "/me/player/recently-played"

func (c Client) RecentlyPlayed(ctx context.Context, limit int) (*RecentlyPlayedResponse, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("%s?limit=%d", recentEndpoint, limit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("recently played: unexpected status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("recently played: unexpected status code: %d", resp.StatusCode)
	}

	var recent RecentlyPlayedResponse
	if err := json.NewDecoder(resp.Body).Decode(&recent); err != nil {
		return nil, err
	}

	return &recent, nil
}

type RecentlyPlayedResponse struct {
	Items []RecentlyPlayedItem `json:"items"`
}

type RecentlyPlayedItem struct {
	Track    Item      `json:"track"`
	PlayedAt time.Time `json:"played_at"`
	Context  Context   `json:"context"`
}
