package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// set up the logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Read the Redis configuration from the environment
	redisAddress := os.Getenv("REDIS_ADDRESS")
	cache := redis.NewClient(&redis.Options{Addr: redisAddress})

	err := cache.Ping(ctx).Err()
	if err != nil {
		slog.Error("error connecting to redis, is REDIS_ADDRESS set?", "error", err, "address", redisAddress)
		os.Exit(1)
	} else {
		slog.Info("connected to redis", "address", redisAddress)
		defer func() {
			if err := cache.Close(); err != nil {
				slog.Error("error closing redis connection", "error", err)
			}
		}()
	}

	// Create a new shorter service
	s := &shorter{cache: cache}

	// Register HTTP handlers
	http.HandleFunc("/manage", s.Manage)
	http.HandleFunc("/", s.Redirect)

	slog.Info("server started", "port", "8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("error starting server", "error", err)
		os.Exit(1)
	}

}

type shorter struct {
	cache *redis.Client
}

// Generate a random url safe string of arbitrary length using crypto/rand.
func randomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

// Manage handles the /manage endpoint that allows users to manage their
// short links.
func (s *shorter) Manage(w http.ResponseWriter, r *http.Request) {
	// POST /manage
	if r.Method == http.MethodPost {
		// create a new short link
		// get the long link from the JSON request body
		var req struct {
			LongLink string `json:"long_link"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "bad request could not decode json", http.StatusBadRequest)
			slog.Error("error decoding JSON request body", "error", err)
			return
		}

		// generate a random short link
		shortLink, err := randomString(10)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			slog.Error("error generating random string for short link", "error", err)
			return
		}
		// save the short link in the cache
		err = s.cache.Set(r.Context(), cacheKey(shortLink), req.LongLink, 7*24*time.Hour).Err()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			slog.Error("error saving short link to cache", "error", err, "short_link", shortLink, "long_link", req.LongLink)
			return
		}
		// return the short link to the user
		resp := struct {
			ShortLink string `json:"short_link"`
		}{ShortLink: shortLink}
		json.NewEncoder(w).Encode(resp)
		return
	}

	http.NotFound(w, r)
}

// construct the redis cache key
func cacheKey(shortLink string) string {
	return "shorter:" + shortLink
}

// Redirect handles the /{shortLink} endpoint that redirects users to the
// long link associated with the short link.
func (s *shorter) Redirect(w http.ResponseWriter, r *http.Request) {
	shortLink := r.URL.Path[1:]
	// check if the short link is empty
	if shortLink == "" {
		http.NotFound(w, r)
		return
	}

	// get the long link from the cache
	longLink, err := s.cache.Get(r.Context(), cacheKey(shortLink)).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		slog.Error("error getting long link from cache", "error", err, "cache_key", cacheKey(shortLink))
		return
	}
	// redirect the user to the long link
	http.Redirect(w, r, longLink, http.StatusFound)
}
