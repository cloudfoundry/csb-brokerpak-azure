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
			fail(w, http.StatusBadRequest, "Key missing.")
			return
		}

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
}
