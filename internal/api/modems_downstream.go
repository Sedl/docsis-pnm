package api

import (
	"database/sql"
	"github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/misc"
	"github.com/sedl/docsis-pnm/internal/types"
	"net/http"
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
	vars, err := ParsePath(r)
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
	query.Where("modem." + vars.ModemColumn, "=", vars.ModemId)
	query.Where("d.poll_time", "=", db.SqlExpression("(SELECT MAX(poll_time) FROM modem_downstream WHERE modem_id = modem.id)"))

	err = query.Exec()
	if err != nil {
		HandleServerError(w, err)
		return
	}
	defer misc.CloseOrLog(query)

	history, err := query.Next()
	if err != nil {
		HandleServerError(w, err)
		return
	}
	hist, _ := history.([]*DownstreamHistory)

	// TODO limit query!
	if hist != nil && len(hist) > 0 {
		JsonResponse(w, hist[0])
		return
	} else {
		JsonResponse(w, nil)
		return
	}

}

func (api *Api) modemsDownstreamHistory(w http.ResponseWriter, r *http.Request) {
    vars, err := ParsePath(r)
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
	query.Where("modem." + vars.ModemColumn, "=", vars.ModemId)
	query.Where("d.poll_time", ">=", vars.FromTs)
	query.Where("d.poll_time", "<=", vars.ToTs)
	defer misc.CloseOrLog(query)
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

	addCacheHeader(vars.ToTs, w)
	hist, _ := history.([]*DownstreamHistory)
	addCountHeader(w, len(hist))
	JsonResponse(w, hist)
}

