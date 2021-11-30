package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleListCollections(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling list collections.")

		databaseName := mux.Vars(r)["database"]
		list, err := client.Database(databaseName).ListCollectionNames(r.Context(), bson.D{})
		if err != nil {
			fail(w, http.StatusNotFound, "error listing collections: %s", err)
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

		log.Printf("Listed collections: %s", strings.Join(list, ", "))
	}
}
