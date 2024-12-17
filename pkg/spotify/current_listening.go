package spotify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shantanuraj/listening/pkg/log"
)

const currentlyListeningEndpoint = "/me/player/currently-playing"

func (c Client) CurrentlyListening(ctx context.Context) (*CurrentlyPlayingResponse, error) {
	resp, err := c.Get(ctx, currentlyListeningEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		log.Errorf("current: unexpected status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("current: unexpected status code: %d", resp.StatusCode)
	}

	var currentlyPlaying CurrentlyPlayingResponse
	if err := json.NewDecoder(resp.Body).Decode(&currentlyPlaying); err != nil {
		log.Errorf("current: failed to decode response: %v", err)
		return nil, err
	}

	return &currentlyPlaying, nil
}

type CurrentlyPlayingResponse struct {
	Timestamp            int64   `json:"timestamp"`
	Device               Device  `json:"device"`
	Context              Context `json:"context"`
	ProgressMS           int64   `json:"progress_ms"`
	Item                 Item    `json:"item"`
	CurrentlyPlayingType string  `json:"currently_playing_type"`
	Actions              Actions `json:"actions"`
	IsPlaying            bool    `json:"is_playing"`
}

type Device struct {
	ID               string `json:"id"`
	IsActive         bool   `json:"is_active"`
	IsPrivateSession bool   `json:"is_private_session"`
	IsRestricted     bool   `json:"is_restricted"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	VolumePercent    int64  `json:"volume_percent"`
	SupportsVolume   bool   `json:"supports_volume"`
}

type Actions struct {
	Disallows Disallows `json:"disallows"`
}

type Disallows struct {
	Resuming bool `json:"resuming"`
}

type Context struct {
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

type ExternalUrls struct {
	Spotify string `json:"spotify"`
}

type Item struct {
	Album        Album        `json:"album"`
	Artists      []Artist     `json:"artists"`
	DiscNumber   int64        `json:"disc_number"`
	DurationMS   int64        `json:"duration_ms"`
	Explicit     bool         `json:"explicit"`
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	IsLocal      bool         `json:"is_local"`
	Name         string       `json:"name"`
	Popularity   int64        `json:"popularity"`
	TrackNumber  int64        `json:"track_number"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

type Album struct {
	AlbumType    string       `json:"album_type"`
	Artists      []Artist     `json:"artists"`
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Images       []Image      `json:"images"`
	Name         string       `json:"name"`
	TotalTracks  int64        `json:"total_tracks"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

type Artist struct {
	ExternalUrls ExternalUrls `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

type Image struct {
	Height int64  `json:"height"`
	URL    string `json:"url"`
	Width  int64  `json:"width"`
}
