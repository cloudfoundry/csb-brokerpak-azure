package app

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handleSet(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		schema, err := schemaName(r)
		if err != nil {
			fail(w, http.StatusInternalServerError, "Schema name error: %s", err)
			return
		}

		key, ok := mux.Vars(r)["key"]
		if !ok {
			fail(w, http.StatusBadRequest, "Key missing.")
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %s.%s (%s, %s) VALUES ($1, $2)`, schema, tableName, keyColumn, valueColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "Error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(key, string(rawValue))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error inserting values: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Key %q set to value %q.", key, string(rawValue))
	}
}
