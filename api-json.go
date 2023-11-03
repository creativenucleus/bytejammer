package main

import (
	"encoding/json"
	"net/http"
)

func apiOutErr(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	type errJson struct {
		Error string `json:"error"`
	}

	out := errJson{
		Error: err.Error(),
	}

	json.NewEncoder(w).Encode(out)
}

func apiOutResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}
