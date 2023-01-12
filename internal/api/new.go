package api

import (
    "github.com/gorilla/mux"
    "github.com/sedl/docsis-pnm/internal/manager"
    "net/http"
)

func NewApi(manager *manager.Manager) *http.Server {
    router := mux.NewRouter().StrictSlash(true)
    Register(router, manager)
    server := &http.Server{
        Addr:    ":8080",
        Handler: LogMiddleware(CacheMiddleware(router)),
    }

    return server
}
