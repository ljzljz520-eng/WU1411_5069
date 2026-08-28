package gateway

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	ResourceID string `json:"resource_id,omitempty"`
	Remaining  int    `json:"remaining,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Error      string `json:"error,omitempty"`
	Code       string `json:"code,omitempty"`
}

func writeJSON(writer http.ResponseWriter, status int, response Response) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(response)
}

func writeError(writer http.ResponseWriter, status int, code, message, requestID string) {
	writeJSON(writer, status, Response{Error: message, Code: code, RequestID: requestID})
}

func statusCode(status int) int {
	if status < 100 || status > 599 {
		return http.StatusInternalServerError
	}
	return status
}
