package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

const playEndpoint = "/me/player/play"

type PlayRequest struct {
	ContextURI string   `json:"context_uri,omitempty"`
	URIs       []string `json:"uris,omitempty"`
	Offset     Offset   `json:"offset,omitempty"`
	PositionMS int      `json:"position_ms,omitempty"`
}

type Offset struct {
	Position int    `json:"position,omitempty"`
	URI      string `json:"uri,omitempty"`
}

func (c Client) Play(ctx context.Context, req PlayRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		log.Printf("play: failed to marshal request: %v", err)
		return fmt.Errorf("play: failed to marshal request: %w", err)
	}
	resp, err := c.Put(ctx, playEndpoint, bytes.NewReader(data))
	if err != nil {
		log.Printf("play: failed to make request: %v", err)
		return err
	}

	if resp.StatusCode != 204 {
		log.Printf("play: unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("play: unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
