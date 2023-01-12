package api

import (
	"github.com/gorilla/mux"
	"github.com/sedl/docsis-pnm/internal/manager"
	"github.com/sedl/docsis-pnm/internal/types"
	"net/http"
)

func NewApi(manager *manager.Manager, cfg *types.ApiConfig) *http.Server {
	router := mux.NewRouter().StrictSlash(true)
	Register(router, manager)
	server := &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: LogMiddleware(CacheMiddleware(router)),
	}

	return server
}
