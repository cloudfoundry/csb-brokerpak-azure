package app

import (
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	documentNameKey = "name"
	documentDataKey = "data"
	documentTTLKey  = "ttl"
)

func App(uri string) http.Handler {
	client := connect(uri)

	r := http.NewServeMux()
	r.HandleFunc("GET /", handleListDatabases(client))
	r.HandleFunc("GET /{database}", handleListCollections(client))
	r.HandleFunc("GET /{database}/{collection}/{document}", handleFetchDocument(client))
	r.HandleFunc("PUT /{database}/{collection}/{document}", handleStoreDocument(client))

	return r
}

func connect(uri string) *mongo.Client {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %s", err)
	}

	return client
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
