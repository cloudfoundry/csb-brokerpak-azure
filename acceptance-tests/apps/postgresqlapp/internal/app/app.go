package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(uri string) *mux.Router {
	db := connect(uri)

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods(http.MethodHead, http.MethodGet)
	r.HandleFunc("/{schema}", handleCreateSchema(db)).Methods(http.MethodPut)
	r.HandleFunc("/{schema}", handleDropSchema(db)).Methods(http.MethodDelete)
	r.HandleFunc("/{schema}/{key}", handleSet(db)).Methods(http.MethodPut)
	r.HandleFunc("/{schema}/{key}", handleGet(db)).Methods(http.MethodGet)

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(uri string) *sql.DB {
	db, err := sql.Open("pgx", uri)
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

func fail(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
