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

		key, ok := mux.Vars(r)["key"]
		if !ok {
			log.Println("Key missing.")
			http.Error(w, "Key missing.", http.StatusBadRequest)
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error parsing value: %s", err)
			http.Error(w, "Failed to parse value.", http.StatusBadRequest)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %s (%s, %s) VALUES (?, ?)`, tableName, keyColumn, valueColumn))
		if err != nil {
			log.Printf("Error preparing statement: %s", err)
			http.Error(w, "Failed to prepare statement.", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(key, string(rawValue))
		if err != nil {
			log.Printf("Error inserting values: %s", err)
			http.Error(w, "Failed to insert values.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Key %q set to value %q.", key, string(rawValue))
	}
}
