package apiutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler/middlewares/auth"
	"github.com/pkg/errors"
)

func SendJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetUserId(r *http.Request) string {
	return r.Header.Get(auth.AuthHeaderUserId)
}

// ReadBody reads bytes from r.Body
// **and closes r.Body** before returning the bytes
func ReadBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CopyBody reads bytes from r.Body,
// and assign a new io.ReadCloser to r.Body
// before returning the bytes read
func CopyBody(r *http.Request) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to close original body")
	}

	body := buf.Bytes()
	l := len(body)
	copied := make([]byte, l)

	n := copy(copied, body)
	if n != l {
		panic(fmt.Errorf("CopyBody error: expecting %d bytes, copied %d bytes", l, n))
	}

	r.Body = io.NopCloser(buf)

	return copied, nil
}
