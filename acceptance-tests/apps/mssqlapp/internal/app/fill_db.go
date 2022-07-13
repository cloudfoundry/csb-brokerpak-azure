package app

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func handleFillDatabase(config string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling fill database.")
		db := connect(config)
		defer db.Close()

		schema, err := schemaName(r)
		if err != nil {
			fail(w, http.StatusInternalServerError, "schema name error: %s", err)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %s.%s (%s, %s) VALUES (@p1, REPLICATE(CAST(@p2 AS VARCHAR(max)), 100000))`, schema, tableName, keyColumn, valueColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		row := randomString()
		log.Printf("inserting row: %s\n", row)
		_, err = stmt.Exec(row, randomString())
		switch {
		case err == nil:
			log.Println("inserted ok")
			w.WriteHeader(http.StatusOK)
		case strings.Contains(err.Error(), "has reached its size quota"):
			log.Println("database full")
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			log.Printf("error inserting into database: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func randomString() string {
	buf := make([]byte, 50)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}
