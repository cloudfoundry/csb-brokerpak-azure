package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(uri string) http.Handler {
	db := connect(uri)

	r := http.NewServeMux()
	r.HandleFunc("GET /", aliveness)
	r.HandleFunc("PUT /{schema}", handleCreateSchema(db))
	r.HandleFunc("DELETE /{schema}", handleDropSchema(db))
	r.HandleFunc("PUT /{schema}/{key}", handleSet(db))
	r.HandleFunc("GET /{schema}/{key}", handleGet(db))

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
	db.SetMaxIdleConns(0)
	return db
}

func schemaName(r *http.Request) (string, error) {
	schema := r.PathValue("schema")

	switch {
	case schema == "":
		return "", fmt.Errorf("schema name must be supplied")
	case len(schema) > 50:
		return "", fmt.Errorf("schema name too long")
	case !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(schema):
		return "", fmt.Errorf("schema name contains invalid characters")
	default:
		return schema, nil
	}
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
