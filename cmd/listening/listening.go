package main

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/shantanuraj/listening/pkg/spotify"
)

var (
	host = cmp.Or(os.Getenv("SL_HOST"), "localhost")
	port = cmp.Or(os.Getenv("SL_PORT"), "5050")
	addr = cmp.Or(os.Getenv("SL_ADDR"), fmt.Sprintf("http://%s:%s", host, port))
)

func main() {
	ctx := context.Background()

	client := spotify.DefaultClient

	mux := http.NewServeMux()

	client.RegisterAuthenticationHandlers(addr, mux)

	mux.HandleFunc("GET /current", func(w http.ResponseWriter, r *http.Request) {
		if !client.IsAuthenticated() {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}

		listening, err := client.CurrentlyListening(ctx)
		if err != nil {
			http.Error(w, "failed to get currently listening", http.StatusInternalServerError)
			return
		}

		if listening == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(listening)
	})

	fmt.Printf("listening on %s\n", addr)
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), mux)
}
