package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handleGet(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")

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

		stmt, err := db.Prepare(fmt.Sprintf(`SELECT %s from %s.%s WHERE %s = $1`, valueColumn, schema, tableName, keyColumn))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(key)
		if err != nil {
			fail(w, http.StatusNotFound, "Error selecting value: %s", err)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			fail(w, http.StatusNotFound, "Error finding value: %s", err)
			return
		}

		var value string
		if err := rows.Scan(&value); err != nil {
			fail(w, http.StatusNotFound, "Error retrieving value: %s", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write([]byte(value))

		if err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}

		log.Printf("Value %q retrived from key %q.", value, key)
	}
}
