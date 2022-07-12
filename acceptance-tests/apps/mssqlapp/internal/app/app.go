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

func App(config string) *mux.Router {
	db := connect(config)
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %s", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	r.HandleFunc("/{schema}", handleCreateSchema(config)).Methods("PUT")
	r.HandleFunc("/{schema}", handleFillDatabase(config)).Methods("POST")
	r.HandleFunc("/{schema}", handleDropSchema(config)).Methods("DELETE")
	r.HandleFunc("/{schema}/{key}", handleSet(config)).Methods("PUT")
	r.HandleFunc("/{schema}/{key}", handleGet(config)).Methods("GET")

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(config string) *sql.DB {
	db, err := sql.Open("sqlserver", config)
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
