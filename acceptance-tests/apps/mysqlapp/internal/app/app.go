package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"
)

func App(config *mysql.Config) *mux.Router {
	db := connect(config)

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	r.HandleFunc("/{key}", handleSet(db)).Methods("PUT")
	r.HandleFunc("/{key}", handleGet(db)).Methods("GET")

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func connect(config *mysql.Config) *sql.DB {
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, tableName, keyColumn, valueColumn))
	if err != nil {
		log.Fatalf("failed to create test table: %s", err)
	}

	return db
}
