package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"postgresqlapp/internal/app"
	"postgresqlapp/internal/credentials"
)

func main() {
	log.Println("Starting.")

	log.Println("Reading credentials.")
	creds, err := credentials.Read()
	if err != nil {
		panic(err)
	}

	port := port()
	log.Printf("Listening on port: %s", port)
	http.Handle("/", app.App(creds))
	http.ListenAndServe(port, nil)
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
