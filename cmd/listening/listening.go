package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/shantanuraj/listening/pkg/funk"
	"github.com/shantanuraj/listening/pkg/log"
	"github.com/shantanuraj/listening/pkg/middleware"
	"github.com/shantanuraj/listening/pkg/spotify"
)

var (
	host           = cmp.Or(os.Getenv("SL_HOST"), "localhost")
	port           = cmp.Or(os.Getenv("SL_PORT"), "5050")
	addr           = cmp.Or(os.Getenv("SL_ADDR"), fmt.Sprintf("http://%s:%s", host, port))
	enabledOrigins = []string{
		cmp.Or(os.Getenv("SL_DEV_ORIGIN"), "http://localhost:4321"),
		cmp.Or(os.Getenv("SL_PROD_ORIGIN"), "https://sraj.me"),
	}
)

type App struct {
	client *spotify.Client
	log    *log.Logger
}

func main() {
	client := spotify.DefaultClient
	log := log.New()
	app := &App{
		client: client,
		log:    log,
	}

	mux := http.NewServeMux()

	client.RegisterAuthenticationHandlers(addr, mux)
	mux.HandleFunc("GET /current", client.AuthMiddleware(app.currentTrackHandler))
	mux.HandleFunc("GET /queue", client.AuthMiddleware(app.queueHandler))
	mux.HandleFunc("GET /recent", client.AuthMiddleware(app.recentHandler))
	mux.HandleFunc("PUT /play", client.AuthMiddleware(app.playHandler))

	enableCors := middleware.WithCors(enabledOrigins)
	enableLogging := middleware.WithLogging(log)

	log.Infof("listening on %s", addr)

	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), enableCors(enableLogging(mux))); err != nil {
		log.Fatalf("failed to listen at %s %v", addr, err)
	}
}

var storedTrack atomic.Value

func (app *App) currentTrackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := app.log
	client := app.client

	query := r.URL.Query()
	skipCache := query.Has("skip-cache")

	writeResponse := true

	if stored := storedTrack.Load(); !skipCache && stored != nil {
		log.Infof("serving stored track")
		track := stored.(*spotify.CurrentlyPlayingResponse)
		writeJSON(w, track)
		writeResponse = false
	}

	log.Infof("fetching currently listening")

	listening, err := client.CurrentlyListening(ctx)
	if err != nil {
		log.Errorf("failed to fetch currently listening: %v", err)
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

	writeJSON(w, listening)
}

var storedQueue atomic.Value

const defaultLimit = 5
const maxLimit = 15

func (app *App) queueHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := app.log
	client := app.client

	query := r.URL.Query()
	skipCache := query.Has("skip-cache")

	limit := defaultLimit
	limitStr := query.Get("limit")
	if limitStr != "" {
		limitValue, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if limitValue < 1 || limitValue > maxLimit {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = limitValue
	}

	writeResponse := true

	if stored := storedQueue.Load(); !skipCache && stored != nil {
		log.Infof("serving stored queue")
		queue := stored.(*spotify.QueueResponse)
		if queue != nil {
			queue.Queue = funk.Range(queue.Queue, 0, limit)
		}
		writeJSON(w, queue)
		writeResponse = false
	}

	log.Infof("fetching queue")

	queue, err := client.Queue(ctx)
	if err != nil {
		log.Errorf("failed to fetch queue: %v", err)
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

	writeJSON(w, queue)
}

func (app *App) recentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := app.log
	client := app.client

	query := r.URL.Query()

	limit := defaultLimit
	limitStr := query.Get("limit")
	if limitStr != "" {
		limitValue, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if limitValue < 1 || limitValue > maxLimit {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = limitValue
	}

	log.Infof("fetching recent tracks")

	recent, err := client.RecentlyPlayed(ctx, limit)
	if err != nil {
		log.Errorf("failed to fetch recently played: %v", err)
		http.Error(w, "failed to fetch recently played", http.StatusInternalServerError)
		return
	}

	writeJSON(w, recent)
}

func (app *App) playHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := app.log
	client := app.client

	var req spotify.PlayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Errorf("play: failed to decode request: %v", err)
		http.Error(w, "play: failed to decode request", http.StatusBadRequest)
		return
	}

	if err := client.Play(ctx, req); err != nil {
		log.Errorf("play: failed to play: %v", err)
		http.Error(w, "play: failed to play", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
