package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	documentNameKey = "name"
	documentDataKey = "data"
	documentTTLKey  = "ttl"
)

func App(uri string) *mux.Router {
	client := connect(uri)

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD")
	r.HandleFunc("/", handleListDatabases(client)).Methods("GET")
	r.HandleFunc("/{database}", handleListCollections(client)).Methods("GET")
	r.HandleFunc("/{database}/{collection}/{document}", handleFetchDocument(client)).Methods("GET")
	r.HandleFunc("/{database}/{collection}/{document}", handleStoreDocument(client)).Methods("PUT")

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
