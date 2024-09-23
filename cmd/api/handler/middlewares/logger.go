package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func marshalJson(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := bytes.NewBuffer(nil)
		_, err := io.Copy(body, r.Body)
		if err != nil {
			log.Fatalf("[clipboard-api] failed to read body: %s", err.Error())
		}

		log.Printf("[clipboard-api] %s %s \"%s\"", r.Method, r.URL, body.Bytes())

		r.Body.Close() //  must close
		// Attach a new body to next handlers
		r.Body = io.NopCloser(body)

		next.ServeHTTP(w, r)
	})
}
