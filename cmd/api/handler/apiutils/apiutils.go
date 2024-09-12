package apiutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler/auth"
)

func SendJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ReadBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GetUserId(r *http.Request) string {
	return r.Header.Get(auth.AuthHeaderUserId)
}
