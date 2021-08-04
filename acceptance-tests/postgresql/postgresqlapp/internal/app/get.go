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

		key, ok := mux.Vars(r)["key"]
		if !ok {
			log.Println("Key missing.")
			http.Error(w, "Key missing.", http.StatusBadRequest)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`SELECT %s from %s WHERE %s = $1`, valueColumn, tableName, keyColumn))
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

		log.Printf("Value %q retrived from key %q.", value, key)
	}
}
