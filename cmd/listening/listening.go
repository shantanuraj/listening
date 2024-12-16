package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

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
	mux.HandleFunc("GET /current", currentTrackHandler(client))

	log.Printf("listening on %s", addr)
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), withCors(mux))
}

var storedTrack atomic.Value

func currentTrackHandler(client *spotify.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		writeResponse := true

		if stored := storedTrack.Load(); stored != nil {
			log.Println("serving stored track")
			track := stored.(*spotify.CurrentlyPlayingResponse)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(track)
			writeResponse = false
		}

		if !client.IsAuthenticated() {
			if client.IsTokenExpired() {
				if err := client.RefreshToken(ctx); err != nil {
					log.Printf("current track: failed to refresh token: %v", err)
					http.Error(w, "current track: failed to refresh token", http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}
		}

		log.Println("fetching currently listening")

		listening, err := client.CurrentlyListening(ctx)
		if err != nil {
			log.Printf("failed to fetch currently listening: %v", err)
		} else {
			storedTrack.Store(listening)
		}
		if !writeResponse {
			return
		}

		if err != nil {
			http.Error(w, "failed to fetch currently listening", http.StatusInternalServerError)
			return
		}

		if listening == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(listening)
	}
}
