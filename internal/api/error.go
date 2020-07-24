package api

import (
    "encoding/json"
    "errors"
    "net/http"
)

var ErrorInvalidModemId = errors.New("invalid ID, MAC or IP address")

func HandleServerError(w http.ResponseWriter, err error) {
    w.WriteHeader(http.StatusInternalServerError)
    resp := ErrorResponse{Message: err.Error()}
    _ = json.NewEncoder(w).Encode(resp)
}

func HandleServerConflict(w http.ResponseWriter, message string) {
    w.WriteHeader(http.StatusConflict)
    resp := ErrorResponse{Message: message}
    _ = json.NewEncoder(w).Encode(resp)
}

func HandleBadRequest(w http.ResponseWriter, err error) {
    w.WriteHeader(http.StatusBadRequest)
    resp := ErrorResponse{Message: err.Error()}
    _ = json.NewEncoder(w).Encode(resp)
}