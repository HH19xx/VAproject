package http

import (
	"encoding/json"
	nethttp "net/http"
	"time"
)

type successEnvelope struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type errorEnvelope struct {
	Success   bool        `json:"success"`
	Error     interface{} `json:"error"`
	Timestamp string      `json:"timestamp"`
}

type apiError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func writeJSON(w nethttp.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func JSONSuccess(w nethttp.ResponseWriter, status int, data interface{}, message string) {
	writeJSON(w, status, successEnvelope{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func JSONError(w nethttp.ResponseWriter, status int, code, message string, details interface{}) {
	writeJSON(w, status, errorEnvelope{
		Success: false,
		Error: apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
