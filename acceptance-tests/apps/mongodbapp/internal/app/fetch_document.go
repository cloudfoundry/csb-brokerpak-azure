package app

import (
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func handleFetchDocument(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling fetch.")

		databaseName := r.PathValue("database")
		if databaseName == "" {
			fail(w, http.StatusBadRequest, "database name must be supplied")
		}
		collectionName := r.PathValue("collection")
		if collectionName == "" {
			fail(w, http.StatusBadRequest, "collection name must be supplied")
		}
		documentName := r.PathValue("document")
		if documentName == "" {
			fail(w, http.StatusBadRequest, "document name must be supplied")
		}

		filter := bson.D{{Key: documentNameKey, Value: documentName}}
		result := client.Database(databaseName).Collection(collectionName).FindOne(r.Context(), filter)
		if result.Err() != nil {
			fail(w, http.StatusNotFound, "error finding document: %s", result.Err())
			return
		}

		var receiver bson.D
		if err := result.Decode(&receiver); err != nil {
			fail(w, http.StatusNotFound, "error decoding document: %s", err)
			return
		}

		var data any
		for _, e := range receiver {
			if e.Key == documentDataKey {
				data = e.Value
			}
		}

		if data == nil {
			fail(w, http.StatusNotFound, "error find document data: %+v", receiver)
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
