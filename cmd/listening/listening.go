package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/shantanuraj/listening/pkg/funk"
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
	mux.HandleFunc("GET /queue", queueHandler(client))
	mux.HandleFunc("PUT /play", playHandler(client))

	log.Printf("listening on %s", addr)
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), withCors(mux))
}

var storedTrack atomic.Value

func currentTrackHandler(client *spotify.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := r.URL.Query()
		skipCache := query.Has("skip-cache")

		writeResponse := true

		if stored := storedTrack.Load(); !skipCache && stored != nil {
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
					http.Error(
						w,
						"current track: failed to refresh token",
						http.StatusInternalServerError,
					)
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

var storedQueue atomic.Value

const defaultQueueLimit = 5
const queueCap = 15

func queueHandler(client *spotify.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := r.URL.Query()
		skipCache := query.Has("skip-cache")

		limit := defaultQueueLimit
		limitStr := query.Get("limit")
		if limitStr != "" {
			limitValue, err := strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			if limitValue < 1 || limitValue > queueCap {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = limitValue
		}

		writeResponse := true

		if stored := storedQueue.Load(); !skipCache && stored != nil {
			log.Println("serving stored queue")
			queue := stored.(*spotify.QueueResponse)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if queue != nil {
				queue.Queue = funk.Range(queue.Queue, 0, limit)
			}
			json.NewEncoder(w).Encode(queue)
			writeResponse = false
		}

		if !client.IsAuthenticated() {
			if client.IsTokenExpired() {
				if err := client.RefreshToken(ctx); err != nil {
					log.Printf("queue: failed to refresh token: %v", err)
					http.Error(
						w,
						"queue: failed to refresh token",
						http.StatusInternalServerError,
					)
					return
				}
			} else {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}
		}

		log.Println("fetching queue")

		queue, err := client.Queue(ctx)
		if err != nil {
			log.Printf("failed to fetch queue: %v", err)
		} else {
			storedQueue.Store(queue)
		}
		if !writeResponse {
			return
		}

		if err != nil {
			http.Error(w, "failed to fetch queue", http.StatusInternalServerError)
			return
		}

		if queue == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		queue.Queue = funk.Range(queue.Queue, 0, limit)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(queue)
	}
}

func playHandler(client *spotify.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if !client.IsAuthenticated() {
			if client.IsTokenExpired() {
				if err := client.RefreshToken(ctx); err != nil {
					log.Printf("play: failed to refresh token: %v", err)
					http.Error(w, "play: failed to refresh token", http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}
		}

		var req spotify.PlayRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("play: failed to decode request: %v", err)
			http.Error(w, "play: failed to decode request", http.StatusBadRequest)
			return
		}

		if err := client.Play(ctx, req); err != nil {
			log.Printf("play: failed to play: %v", err)
			http.Error(w, "play: failed to play", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
