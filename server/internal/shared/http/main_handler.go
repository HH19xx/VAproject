package http

import (
	"encoding/json"
	nethttp "net/http"
)

type Message struct {
	Message string `json:"message"`
}

func Handler(w nethttp.ResponseWriter, r *nethttp.Request) {
	response := Message{Message: "こんにちは、みなさん"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
	}
}
