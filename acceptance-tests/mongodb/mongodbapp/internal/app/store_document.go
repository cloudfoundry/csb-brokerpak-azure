package app

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleStoreDocument(client *mongo.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling store.")
		databaseName := mux.Vars(r)["database"]
		collectionName := mux.Vars(r)["collection"]
		documentName := mux.Vars(r)["document"]

		rawData, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error parsing data: %s", err)
			http.Error(w, "Failed to parse data.", http.StatusBadRequest)
			return
		}

		data := string(rawData)
		document := bson.M{documentNameKey: documentName, documentDataKey: data}

		result, err := client.Database(databaseName).Collection(collectionName).InsertOne(r.Context(), document)
		if err != nil {
			log.Printf("Error creating document %q with data %q in database %q, collection %q: %s", documentName, data, databaseName, collectionName, err)
			http.Error(w, "Failed to create document.", http.StatusFailedDependency)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Created document %q with data %q in database %q, collection %q.", result.InsertedID, documentName, data, databaseName, collectionName)
	}
}
