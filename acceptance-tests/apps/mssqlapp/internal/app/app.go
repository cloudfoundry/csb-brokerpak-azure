package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/mux"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(uri string) *mux.Router {
	db := connect(uri)

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	r.HandleFunc("/{schema}", handleCreateSchema(db)).Methods("PUT")
	r.HandleFunc("/{schema}", handleDropSchema(db)).Methods("DELETE")
	r.HandleFunc("/{schema}/{key}", handleSet(db)).Methods("PUT")
	r.HandleFunc("/{schema}/{key}", handleGet(db)).Methods("GET")

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(uri string) *sql.DB {
	db, err := sql.Open("sqlserver", uri)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}

	return db
}

func schemaName(r *http.Request) (string, error) {
	schema, ok := mux.Vars(r)["schema"]

	switch {
	case !ok:
		return "", fmt.Errorf("schema missing")
	case len(schema) > 50:
		return "", fmt.Errorf("schema name too long")
	case len(schema) == 0:
		return "", fmt.Errorf("schema name cannot be zero length")
	case !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(schema):
		return "", fmt.Errorf("schema name contains invalid characters")
	default:
		return schema, nil
	}
}
