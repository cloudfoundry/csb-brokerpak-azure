package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func handleListDatabases(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling list database.")
		list, err := client.ListDatabaseNames(r.Context(), bson.D{})
		if err != nil {
			fail(w, http.StatusNotFound, "error listing databases: %s", err)
			return
		}

		data, err := json.Marshal(list)
		if err != nil {
			fail(w, http.StatusNotFound, "JSON error: %s", err)
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
