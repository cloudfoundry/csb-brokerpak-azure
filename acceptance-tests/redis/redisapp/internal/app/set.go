package app

import (
	"io"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func handleSet(client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		key, ok := mux.Vars(r)["key"]
		if !ok {
			log.Println("Key missing.")
			http.Error(w, "Key missing.", http.StatusBadRequest)
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error parsing value: %s", err)
			http.Error(w, "Failed to parse value.", http.StatusBadRequest)
			return
		}

		value := string(rawValue)
		if err := client.Set(r.Context(), key, value, 0).Err(); err != nil {
			log.Printf("Error setting key %q to value %q: %s", key, value, err)
			http.Error(w, "Failed to set value.", http.StatusFailedDependency)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Key %q set to value %q.", key, value)
	}
}
