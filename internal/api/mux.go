package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sedl/docsis-pnm/internal/manager"
	"net/http"
)

type Api struct {
	Manager *manager.Manager
}

type ErrorResponse struct {
	Message string `json:"message"`
}

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

func JsonResponse(w http.ResponseWriter, jsonobj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(jsonobj)
	if err != nil {
		HandleServerError(w, err)
	}
}

func Register(router *mux.Router, manager *manager.Manager) {

	api := Api{Manager: manager}
	router.HandleFunc("/health/status", api.healthStatus)
	router.HandleFunc("/stats", api.stats)

	router.HandleFunc("/cmts", api.cmtsList).Methods("GET")
	router.HandleFunc("/cmts", api.cmtsCreate).Methods("POST")
	router.HandleFunc("/cmts/{cmtsId:[0-9]+}", api.cmtsOne)
	router.HandleFunc("/cmts/{cmtsId:[0-9]+}/modems", api.modemsByCmtsId)

	router.HandleFunc("/modems", api.modemsAll)

	// first test for integer
	router.HandleFunc("/modems/{modemId:[0-9]+$}", api.modemsById)
	// test for IP
	router.HandleFunc("/modems/{modemIp:[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+$}", api.modemsByIp)
	// then test for MAC address
	router.HandleFunc("/modems/{modemMac}", api.modemsByMac)

	router.HandleFunc("/modems/{modemMac}/downstream/history/{fromTS:[0-9]+}/{toTS:[0-9]+}", api.modemsDownstreamByDay )
}
