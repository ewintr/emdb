package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Index(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"message":"emdb index"}`)
}

func Error(w http.ResponseWriter, status int, message string, err error) {
	w.WriteHeader(status)

	var resBody []byte
	res := struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}{
		Message: message,
		Error:   err.Error(),
	}
	resBody, _ = json.Marshal(res)

	fmt.Fprint(w, string(resBody))
}
