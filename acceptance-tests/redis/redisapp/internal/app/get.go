package app

import (
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func handleGet(client *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")

		key, ok := mux.Vars(r)["key"]
		if !ok {
			log.Println("Key missing.")
			http.Error(w, "Key missing.", http.StatusBadRequest)
			return
		}

		value, err := client.Get(r.Context(), key).Result()
		if err != nil {
			log.Printf("Error retrieving value: %s", err)
			http.Error(w, "Failed to retrieve value.", http.StatusNotFound)
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
}
