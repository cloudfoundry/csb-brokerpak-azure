package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/denisenkom/go-mssqldb"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(config string) http.Handler {
	db := connect(config)
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %s", err)
	}

	r := http.NewServeMux()
	r.HandleFunc("GET /", aliveness)
	r.HandleFunc("PUT /{schema}", handleCreateSchema(config))
	r.HandleFunc("POST /{schema}", handleFillDatabase(config))
	r.HandleFunc("DELETE /{schema}", handleDropSchema(config))
	r.HandleFunc("PUT /{schema}/{key}", handleSet(config))
	r.HandleFunc("GET /{schema}/{key}", handleGet(config))

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
