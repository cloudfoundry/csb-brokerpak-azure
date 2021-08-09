package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handleGet(config string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")
		db := connect(config)
		defer db.Close()

		schema, err := schemaName(r)
		if err != nil {
			log.Printf("Schema name error: %s\n", err)
			http.Error(w, "Schema name error.", http.StatusInternalServerError)
			return
		}

		key, ok := mux.Vars(r)["key"]
		if !ok {
			log.Println("Key missing.")
			http.Error(w, "Key missing.", http.StatusBadRequest)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`SELECT %s from %s.%s WHERE %s = @p1`, valueColumn, schema, tableName, keyColumn))
		if err != nil {
			log.Printf("Error preparing statement: %s", err)
			http.Error(w, "Failed to prepare statement.", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(key)
		if err != nil {
			log.Printf("Error selecting value: %s", err)
			http.Error(w, "Failed to select value.", http.StatusNotFound)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			log.Printf("Error finding value: %s", err)
			http.Error(w, "Failed to find value.", http.StatusNotFound)
			return
		}

		var value string
		if err := rows.Scan(&value); err != nil {
			log.Printf("Error retrieving value: %s", err)
			http.Error(w, "Failed to retrieve value.", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write([]byte(value))

		if err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}

		log.Printf("Value %q retrived from key %q in schema %s.", value, key, schema)
	}
}
