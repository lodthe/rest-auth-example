package restapi

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Result interface{}    `json:"result,omitempty"`
	Error  *ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func writeError(w http.ResponseWriter, msg string, code int) {
	writeResponse(w, &Response{
		Error: &ErrorResponse{
			Message: msg,
			Code:    code,
		},
	})
}

func writeResult(w http.ResponseWriter, result interface{}) {
	writeResponse(w, &Response{
		Result: result,
	})
}

func writeResponse(w http.ResponseWriter, resp *Response) {
	if resp.Error != nil {
		w.WriteHeader(resp.Error.Code)
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("failed to encode response: %v\nresponse:\n%v\n\n", err, resp)
	}
}
