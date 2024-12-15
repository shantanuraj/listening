package main

import (
	"context"
	"fmt"
	"log"

	"github.com/shantanuraj/listening/pkg/spotify"
)

func main() {
	ctx := context.Background()

	client := spotify.DefaultClient

	if !client.IsAuthenticated() {
		token, err := client.Authenticate(ctx)
		if err != nil {
			log.Fatalf("failed to authenticate: %v", err)
		}
		client.SetToken(token)
	}

	listening, err := client.CurrentlyListening(ctx)
	if err != nil {
		log.Fatalf("failed to get currently listening: %v", err)
	}

	if listening == nil {
		fmt.Println("not listening to anything")
		return
	}

	fmt.Printf("listening to %s by %s\n", listening.Item.Name, listening.Item.Artists[0].Name)
}
