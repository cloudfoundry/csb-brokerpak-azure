package main

import (
	"fmt"
	"log"
	"mssqlapp/internal/app"
	"mssqlapp/internal/credentials"
	"net/http"
	"os"
)

func main() {
	log.Println("Starting.")

	log.Println("Reading credentials.")
	config, err := credentials.Read()
	if err != nil {
		log.Fatalf("failed to read credentials: %s", err)
	}

	port := port()
	log.Printf("Listening on port: %s", port)
	http.Handle("/", app.App(config))
	http.ListenAndServe(port, nil)
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
