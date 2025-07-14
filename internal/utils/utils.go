package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	bodyToSend, _ := json.Marshal(errorResponse{
		Error: fmt.Sprintf("%s", msg),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(bodyToSend)
	return
}


func RespondWithJSon(w http.ResponseWriter, code int, payload interface{}) error {
	bodyToSend, err := json.Marshal(&payload)
	
	if err != nil {
		return nil
	}

	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(bodyToSend)

	return nil
}
