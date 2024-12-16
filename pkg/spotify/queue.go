package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

const queueEndpoint = "/me/player/queue"

func (c Client) Queue(ctx context.Context) (*QueueResponse, error) {
	resp, err := c.Get(ctx, queueEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		log.Printf("queue: unexpected status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("queue: unexpected status code: %d", resp.StatusCode)
	}

	var queue QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queue); err != nil {
		return nil, err
	}

	return &queue, nil
}

type QueueResponse struct {
	// We intend to use currently_playing from the current endpoint instead
	// CurrentlyPlaying *QueueItem  `json:"currently_playing"`
	Queue            []QueueItem `json:"queue"`
}

type QueueItem struct {
	Album            Album        `json:"album"`
	Artists          []Artist     `json:"artists"`
	DiscNumber       int          `json:"disc_number"`
	DurationMs       int          `json:"duration_ms"`
	Explicit         bool         `json:"explicit"`
	ExternalURLs     ExternalUrls `json:"external_urls"`
	Href             string       `json:"href"`
	ID               string       `json:"id"`
	IsPlayable       bool         `json:"is_playable"`
	Name             string       `json:"name"`
	Popularity       int          `json:"popularity"`
	PreviewURL       string       `json:"preview_url"`
	TrackNumber      int          `json:"track_number"`
	Type             string       `json:"type"`
	URI              string       `json:"uri"`
	IsLocal          bool         `json:"is_local"`
}
