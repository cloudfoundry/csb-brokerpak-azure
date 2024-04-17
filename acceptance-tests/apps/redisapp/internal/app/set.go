package app

import (
	"io"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func handleSet(w http.ResponseWriter, r *http.Request, key string, client *redis.Client) {
	log.Println("Handling set.")

	rawValue, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
		http.Error(w, "Failed to parse value.", http.StatusBadRequest)
		return
	}

	value := string(rawValue)
	if err := client.Set(r.Context(), key, value, 0).Err(); err != nil {
		fail(w, http.StatusFailedDependency, "Error setting key %q to value %q: %s", key, value, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Key %q set to value %q.", key, value)
}
