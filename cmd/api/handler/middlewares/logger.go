package middlewares

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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

type logOutputV1 struct {
	prefix  string
	Time    time.Time         `json:"time"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
}

func (l *logOutputV1) string() string {
	return fmt.Sprintf("[%s] %s %s %v", l.prefix, l.Method, l.URL, l.Headers)
}

func (l *logOutputV1) stringWithBody() string {
	return fmt.Sprintf("[%s] %s %s %v \"%s\"", l.prefix, l.Method, l.URL, l.Headers, string(l.Body))
}

func (l *logOutputV1) MarshalJSON() ([]byte, error) {
	return json.Marshal(l)
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
			now := time.Now()

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

			output := logOutputV1{
				prefix:  prefix,
				Time:    now,
				Method:  r.Method,
				URL:     r.URL.String(),
				Headers: headers,
				Body:    string(body),
			}

			if !jsonOutput {
				s := output.string()
				if logBody {
					s = output.stringWithBody()
				}

				log.Println(s)

				next.ServeHTTP(w, r)
				return
			}

			fmt.Println(marshalJson(output, prefix))
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
