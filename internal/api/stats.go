package api

import "net/http"


func (api *Api) stats(w http.ResponseWriter, r* http.Request) {
	stats := api.Manager.Stats()
	JsonResponse(w, stats)
}