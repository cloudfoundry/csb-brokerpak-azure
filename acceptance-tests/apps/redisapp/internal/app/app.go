package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
)

func App(options *redis.Options) http.HandlerFunc {
	client := redis.NewClient(options)

	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodGet:
			handleGet(w, r, key, client)
		case http.MethodPut:
			handleSet(w, r, key, client)
		default:
			fail(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}
	}
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
