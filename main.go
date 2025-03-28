package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// read the redis configuration from the environment
	redisAddress := os.Getenv("REDIS_ADDRESS")
	cache := redis.NewClient(&redis.Options{Addr: redisAddress})
	log.Println("CACHE IS:", cache)

	// create a new shorter service
	s := &shorter{cache: cache}

	http.HandleFunc("POST /manage", s.Manage)
	http.HandleFunc("GET /", s.Redirect)
	log.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
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
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
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
			return
		}

		// generate a random short link
		shortLink, err := randomString(10)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// save the short link in the cache
		err = s.cache.Set(r.Context(), cacheKey(shortLink), req.LongLink, 7*24*time.Hour).Err()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// return the short link to the user
		resp := struct {
			ShortLink string `json:"short_link"`
		}{ShortLink: shortLink}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

// construct the redis cache key
func cacheKey(shortLink string) string {
	return "shorter:" + shortLink
}

// Redirect handles the /{shortLink} endpoint that redirects users to the
// long link associated with the short link.
func (s *shorter) Redirect(w http.ResponseWriter, r *http.Request) {
	// GET /{shortLink}
	shortLink := r.URL.Path[1:]
	// get the long link from the cache
	longLink, err := s.cache.Get(r.Context(), cacheKey(shortLink)).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// redirect the user to the long link
	http.Redirect(w, r, longLink, http.StatusFound)
}
