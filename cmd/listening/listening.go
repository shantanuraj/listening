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
