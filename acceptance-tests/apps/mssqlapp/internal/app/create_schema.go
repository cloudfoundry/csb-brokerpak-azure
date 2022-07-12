package app

import (
	"fmt"
	"log"
	"net/http"
)

func handleCreateSchema(config string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling create schema.")
		db := connect(config)
		defer db.Close()

		schema, err := schemaName(r)
		if err != nil {
			fail(w, http.StatusInternalServerError, "schema name error: %s", err)
			return
		}

		statement := fmt.Sprintf(`CREATE SCHEMA %s`, schema)
		switch r.URL.Query().Get("dbo") {
		case "", "true":
			statement = statement + " AUTHORIZATION dbo"
		case "false":
		default:
			fail(w, http.StatusBadRequest, "invalid value for dbo")
			return
		}

		if _, err = db.Exec(statement); err != nil {
			fail(w, http.StatusBadRequest, "failed to create schema: %s", err)
			return
		}

		if _, err = db.Exec(fmt.Sprintf(`CREATE TABLE %s.%s (%s VARCHAR(255) NOT NULL, %s VARCHAR(max) NOT NULL)`, schema, tableName, keyColumn, valueColumn)); err != nil {
			fail(w, http.StatusBadRequest, "error creating table: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Schema %q created", schema)
	}
}
