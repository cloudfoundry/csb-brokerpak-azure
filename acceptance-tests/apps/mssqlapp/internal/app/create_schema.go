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
			log.Printf("Schema name error: %s\n", err)
			http.Error(w, "Schema name error.", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`CREATE SCHEMA %s`, schema))
		if err != nil {
			log.Printf("Error creating schema: %s", err)
			http.Error(w, "Failed to create schema.", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`CREATE TABLE %s.%s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, schema, tableName, keyColumn, valueColumn))
		if err != nil {
			log.Printf("Error creating table: %s", err)
			http.Error(w, "Failed to create table.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Schema %q created", schema)
	}
}
