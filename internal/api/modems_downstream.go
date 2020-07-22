package api

import (
	"github.com/gorilla/mux"
	"github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/types"
	"net"
	"net/http"
	"strconv"
)

// TODO implement caching using cache headers
const cacheOffset = 3600

type DownstreamHistory struct {
	TS int64 `json:"ts"`
	DS *[]*types.DownstreamChannel `json:"ds"`
}


func (api *Api)modemsDownstreamByDay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	mac := vars["modemMac"]
	hwaddr, err := net.ParseMAC(mac)
	if err != nil {
		// invalid MAC address
		// we don't care, just return a 404
		w.WriteHeader(http.StatusNotFound)
		return
	}

	from, err := strconv.ParseInt(vars["fromTS"], 10, 64)
	if err != nil {
		HandleServerError(w, err)
		return
	}

	to, err := strconv.ParseInt(vars["toTS"], 10, 64)
	if err != nil {
		HandleServerError(w, err)
		return
	}

	db1, err := api.Manager.GetDbInterface().GetConn()
	if err != nil {
		HandleServerError(w, err)
		return
	}

	dsquery, err := db.NewDownstreamQuery(db1, &hwaddr, "poll_time >= $2 AND poll_time <= $3", from, to)
	if err != nil {
		HandleServerError(w, err)
		return
	}

	// channels := make([]*DownstreamChannel, 0)
	history := make([]*DownstreamHistory, 0)

	var lastTS int64
	dslist := make([]*types.DownstreamChannel, 0)
	for {
		ts, ds, err := dsquery.Next()
		if err != nil {
			HandleServerError(w, err)
			return
		}
		if ds == nil {
			break
		}

		if lastTS != ts {
			dslist = make([]*types.DownstreamChannel, 0)
			history = append(history, &DownstreamHistory{
				TS: ts,
				DS: &dslist,
			})
			lastTS = ts
		}
		dslist = append(dslist, ds)
	}

	w.Header().Set("X-Count", strconv.Itoa(len(history)))
	JsonResponse(w, history)
}

