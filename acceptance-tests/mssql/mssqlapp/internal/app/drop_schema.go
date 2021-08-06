package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

func handleDropSchema(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling drop schema.")

		schema, err := schemaName(r)
		if err != nil {
			log.Printf("Schema name error: %s\n", err)
			http.Error(w, "Schema name error.", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`DROP TABLE %s.%s`, schema, tableName))
		if err != nil {
			log.Printf("Error creating table: %s", err)
			http.Error(w, "Failed to create table.", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`DROP SCHEMA %s`, schema))
		if err != nil {
			log.Printf("Error creating schema: %s", err)
			http.Error(w, "Failed to drop schema.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("Schema %q dropped", schema)
	}
}
