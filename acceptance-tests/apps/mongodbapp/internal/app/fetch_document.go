package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleFetchDocument(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling fetch.")

		databaseName := mux.Vars(r)["database"]
		collectionName := mux.Vars(r)["collection"]
		documentName := mux.Vars(r)["document"]

		filter := bson.D{{Key: documentNameKey, Value: documentName}}
		result := client.Database(databaseName).Collection(collectionName).FindOne(r.Context(), filter)
		if result.Err() != nil {
			log.Printf("error finding document: %s", result.Err())
			http.Error(w, "Failed to finding document.", http.StatusNotFound)
			return
		}

		var receiver bson.D
		if err := result.Decode(&receiver); err != nil {
			log.Printf("error decoding document: %s", err)
			http.Error(w, "Failed to decode document.", http.StatusNotFound)
			return
		}

		var data interface{}
		for _, e := range receiver {
			if e.Key == documentDataKey {
				data = e.Value
			}
		}

		if data == nil {
			log.Printf("error find document data: %+v", receiver)
			http.Error(w, "Failed to find document data.", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err := w.Write([]byte(fmt.Sprintf("%v", data)))
		if err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}

		log.Printf("Data %q retrived from document %q.", data, documentName)
	}
}
