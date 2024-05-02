package app

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
)

func handleSet(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		schema, err := schemaName(r)
		if err != nil {
			fail(w, http.StatusInternalServerError, "Schema name error: %s", err)
			return
		}

		key := r.PathValue("key")
		if key == "" {
			fail(w, http.StatusBadRequest, "key name must be supplied")
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
			return
		}

		if schema == "public" {
			_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, tableName, keyColumn, valueColumn))
			if err != nil {
				log.Fatalf("failed to create test table: %s", err)
			}
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
