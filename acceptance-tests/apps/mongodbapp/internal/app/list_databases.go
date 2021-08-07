package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleListDatabases(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling list database.")
		list, err := client.ListDatabaseNames(r.Context(), bson.D{})
		if err != nil {
			log.Printf("error listing databases: %s", err)
			http.Error(w, "Failed to list databases.", http.StatusNotFound)
			return
		}

		data, err := json.Marshal(list)
		if err != nil {
			log.Printf("JSON error: %s", err)
			http.Error(w, "Failed to serialize database list.", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write(data)
		if err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}

		log.Printf("Listed database: %s", strings.Join(list, ", "))
	}
}
