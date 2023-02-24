package app

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

func handleGet(w http.ResponseWriter, r *http.Request, key string, client *redis.Client) {
	log.Println("Handling get.")

	value, err := client.Get(r.Context(), key).Result()
	if err != nil {
		fail(w, http.StatusNotFound, "Error retrieving value: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(value))

	if err != nil {
		log.Printf("Error writing value: %s", err)
		return
	}

	log.Printf("Value %q retrived from key %q.", value, key)
}
