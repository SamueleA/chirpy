package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

type CorrectedStringReturn struct {
	CorrectedMsg string	
	WasCensored bool
} 

func GetCorrectedString(msg string, prohibitedWords []string) CorrectedStringReturn {
	splitMsg := strings.Split(msg, " ")

	wasCensored := false

	for i := range(len(splitMsg)) {
		for j := range(len(prohibitedWords)) {
			if strings.EqualFold(splitMsg[i], prohibitedWords[j]) {
				splitMsg[i] = "****"
				wasCensored = true
			}
		}
	}

	newString := strings.Join(splitMsg, " ")

	return CorrectedStringReturn {
		CorrectedMsg: newString,
		WasCensored: wasCensored,
	}
}