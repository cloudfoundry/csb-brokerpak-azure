package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	documentNameKey = "name"
	documentDataKey = "data"
	documentTTLKey  = "ttl"
)

func App(uri string) http.Handler {
	client := connect(uri)

	r := chi.NewRouter()
	r.Head("/", aliveness)
	r.Get("/", handleListDatabases(client))
	r.Get("/{database}", handleListCollections(client))
	r.Get("/{database}/{collection}/{document}", handleFetchDocument(client))
	r.Put("/{database}/{collection}/{document}", handleStoreDocument(client))

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(uri string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
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
