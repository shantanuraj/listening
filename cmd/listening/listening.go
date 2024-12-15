package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log"
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
	client := spotify.DefaultClient

	mux := http.NewServeMux()

	client.RegisterAuthenticationHandlers(addr, mux)

	mux.HandleFunc("GET /current", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if !client.IsAuthenticated() {
			if client.IsTokenExpired() {
				if err := client.RefreshToken(ctx); err != nil {
					http.Error(w, "failed to refresh token", http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}
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

	log.Printf("listening on %s", addr)
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), mux)
}
