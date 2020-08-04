package api

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/types"
	"net/http"
	"strconv"
)

const modemHistoryQuery = `
SELECT
    d.poll_time,
    d.freq,
    d.power,
    d.snr,
    d.microrefl,
    d.unerroreds,
    d.correcteds,
    d.erroreds,
    d.modulation 
FROM
    modem_downstream AS d 
JOIN
    modem 
        ON (
            modem.id = d.modem_id
        ) 
`

// TODO implement caching using cache headers
const cacheOffset = 3600

type DownstreamHistory struct {
	TS int64 `json:"ts"`
	DS []*types.DownstreamChannel `json:"ds"`
}

func dsHistoryCb(rows *sql.Rows) (interface{}, error) {

	history := make([]*DownstreamHistory, 0)

	var lastTS, ts int64
	var err error

	var hist *DownstreamHistory
	for rows.Next() {

		ds := &types.DownstreamChannel{}
		err = rows.Scan(
			&ts,
			&ds.Freq,
			&ds.Power,
			&ds.SNR,
			&ds.Microrefl,
			&ds.Unerroreds,
			&ds.Correcteds,
			&ds.Uncorrectables,
			&ds.Modulation,
		)
		if err != nil {
			return nil, err
		}

		if hist == nil || lastTS != ts {
			// dslistNew := make([]*types.DownstreamChannel, 0)
			hist = &DownstreamHistory{
				TS: ts,
				DS: make([]*types.DownstreamChannel, 0),
			}
			history = append(history, hist)
			lastTS = ts
			// dslist = &hist.DS
		}
		hist.DS = append(hist.DS, ds)

	}
	return history, nil
}

func (api *Api) modemsDownstreamLatest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["modemId"]
	column, err := detectModemIdUrlColumn(id)
	if err != nil {
		HandleBadRequest(w, err)
		return
	}

	conn, err := api.Manager.GetDbInterface().GetConn()
	if err != nil {
		HandleServerError(w, err)
		return
	}

	query := db.NewQuery(conn, dsHistoryCb, modemHistoryQuery)
	query.OrderBy("d.poll_time").Where("modem." + column, "=", id)

	err = query.Exec()
	if err != nil {
		HandleServerError(w, err)
		return
	}
	defer db.CloseOrLog(query)

	history, err := query.Next()
	if err != nil {
		HandleServerError(w, err)
		return
	}
	hist, _ := history.([]*DownstreamHistory)

	if hist != nil && len(hist) > 0 {
		JsonResponse(w, hist[0])
		return
	} else {
		JsonResponse(w, nil)
		return
	}

}

func (api *Api) modemsDownstreamHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["modemId"]
	column, err := detectModemIdUrlColumn(id)
	if err != nil {
		HandleBadRequest(w, err)
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

	conn, err := api.Manager.GetDbInterface().GetConn()
	if err != nil {
		HandleServerError(w, err)
		return
	}

	query := db.NewQuery(conn, dsHistoryCb, modemHistoryQuery)
	query.Where("modem." + column, "=", id)
	query.Where("d.poll_time", ">=", from)
	query.Where("d.poll_time", "<=", to)
	err = query.Exec()
	if err != nil {
		HandleServerError(w, err)
		return
	}

	history, err := query.Next()
	if err != nil {
		HandleServerError(w, err)
		return
	}
	hist, _ := history.([]*DownstreamHistory)

	w.Header().Set("X-Count", strconv.Itoa(len(hist)))
	JsonResponse(w, hist)
}

