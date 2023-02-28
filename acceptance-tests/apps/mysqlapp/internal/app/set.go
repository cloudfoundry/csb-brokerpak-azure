package app

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
)

func handleSet(w http.ResponseWriter, r *http.Request, key string, db *sql.DB) {
	log.Println("Handling set.")

	rawValue, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
		return
	}

	stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %s (%s, %s) VALUES (?, ?)`, tableName, keyColumn, valueColumn))
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
