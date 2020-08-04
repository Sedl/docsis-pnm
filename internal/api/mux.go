package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sedl/docsis-pnm/internal/manager"
	"net/http"
	"time"
)

type Api struct {
	Manager *manager.Manager
}

type ErrorResponse struct {
	Message string `json:"message"`
}

const cacheOffset = 3600

func addCacheHeader(timestamp int64, w http.ResponseWriter) {
	now := time.Now().Unix()
	if now - cacheOffset > timestamp {
		w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
	}
}

func JsonResponse(w http.ResponseWriter, jsonobj interface{}) {
	w.Header().Set("Content-Type", "application/json")
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

	router.HandleFunc("/modems/{modemId}", api.modemsById)

	router.HandleFunc("/modems/{modemId}/downstream/latest", api.modemsDownstreamLatest)
	router.HandleFunc("/modems/{modemId}/downstream/history/{fromTS:[0-9]+}/{toTS:[0-9]+}", api.modemsDownstreamHistory)

	router.HandleFunc("/modems/{modemId}/upstream/latest", api.modemsUpstreamLatest)
	router.HandleFunc("/modems/{modemId}/upstream/history/{fromTS:[0-9]+}/{toTS:[0-9]+}", api.modemsUpstreamHistory)
}
