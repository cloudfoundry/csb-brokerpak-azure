package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

func handleGet(w http.ResponseWriter, r *http.Request, key string, db *sql.DB) {
	log.Println("Handling get.")

	stmt, err := db.Prepare(fmt.Sprintf(`SELECT %s from %s WHERE %s = ?`, valueColumn, tableName, keyColumn))
	if err != nil {
		fail(w, http.StatusInternalServerError, "Error preparing statement: %s", err)
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
