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
			fail(w, http.StatusInternalServerError, "Schema name error: %s\n", err)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`DROP TABLE %s.%s`, schema, tableName))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error dropping table: %s", err)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`DROP SCHEMA %s`, schema))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error creating schema: %s", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("Schema %q dropped", schema)
	}
}
