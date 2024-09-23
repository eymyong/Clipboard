package middlewares

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler/apiutils"
)

const defaultPrefix = "clipboard-api"

func DefaultLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := copyBody(r, defaultPrefix)
		log.Printf("[%s] %s %s \"%s\"", defaultPrefix, r.Method, r.URL, body)
		next.ServeHTTP(w, r)
	})
}

// NewLoggerV1 returns a new middleware (mux.MiddlewareFunc)
func NewLoggerV1(
	prefix string,
	logHeaders []string,
	logBody bool,
	jsonOutput bool,
) func(http.Handler) http.Handler {

	if prefix == "" {
		prefix = defaultPrefix
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body []byte
			var err error

			if logBody {
				body, err = apiutils.CopyBody(r)
				if err != nil {
					log.Fatalf("[%s] unexpected error when copying body", prefix)
				}
			}

			headers := make(map[string]string)
			for _, h := range logHeaders {
				headers[h] = r.Header.Get(h)
			}

			if !jsonOutput {
				if logBody {
					log.Printf("[%s] %s %s %v \"%s\"", prefix, r.Method, r.URL, headers, body)
				} else {
					log.Printf("[%s] %s %s %v", prefix, r.Method, r.URL, headers)
				}

				next.ServeHTTP(w, r)
				return
			}

			m := map[string]interface{}{
				"url":     fmt.Sprintf("%s %s", r.Method, r.URL),
				"headers": headers,
			}

			if logBody {
				m["body"] = string(body)
			}

			fmt.Println(marshalJson(m, prefix))

			next.ServeHTTP(w, r)
		})
	}
}

func copyBody(r *http.Request, prefix string) []byte {
	body, err := apiutils.CopyBody(r)
	if err != nil {
		log.Fatalf("[%s] unexpected error when copying body: %s", prefix, err.Error())
	}

	return body
}

func marshalJson(data interface{}, prefix string) string {
	b, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("[%s] unexpected error marshaling log output: %s", prefix, err.Error())
	}

	return string(b)
}
