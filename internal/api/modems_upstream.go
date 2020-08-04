package api

import (
    "database/sql"
    "github.com/gorilla/mux"
    "github.com/sedl/docsis-pnm/internal/db"
    "github.com/sedl/docsis-pnm/internal/types"
    "net/http"
    "strconv"
)

const modemUpstreamHistoryQuery = `
SELECT
    us.poll_time,
    us.modem_id,
    us.us_id,
    us.power_rx,
    us.snr,
    us.microrefl,
    us.unerroreds,
    us.correcteds,
    us.erroreds
FROM
    cmts_upstream_history_modem AS us
JOIN
    modem 
        ON (
            modem.id = us.modem_id
        ) 
`
type UpstreamHistory struct {
    TS int64 `json:"ts"`
    US []*types.UpstreamModemCMTS `json:"us"`
}
func usHistoryCb(rows *sql.Rows) (interface{}, error) {

    history := make([]*UpstreamHistory, 0)

    var lastTS, ts int64
    var err error
    var hist *UpstreamHistory

    for rows.Next() {

        us := &types.UpstreamModemCMTS{}
        err = rows.Scan(
            &ts,
            &us.ModemId,
            &us.UpstreamId,
            &us.PowerRx,
            &us.SNR,
            &us.Microrefl,
            &us.Unerroreds,
            &us.Correcteds,
            &us.Erroreds,
        )
        if err != nil {
            return nil, err
        }

        if hist == nil || lastTS != ts {
            // dslistNew := make([]*types.DownstreamChannel, 0)
            hist = &UpstreamHistory{
                TS: ts,
                US: make([]*types.UpstreamModemCMTS, 0),
            }
            history = append(history, hist)
            lastTS = ts
            // dslist = &hist.DS
        }
        hist.US = append(hist.US, us)
    }
    return history, nil
}

func (api *Api) modemsUpstreamLatest(w http.ResponseWriter, r *http.Request) {

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
    query := db.NewQuery(conn, usHistoryCb, modemUpstreamHistoryQuery)
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

    hist, _ := history.([]*UpstreamHistory)

    if hist != nil && len(hist) > 0 {
        JsonResponse(w, hist[0])
    } else {
        JsonResponse(w, nil)
        return
    }
}

func (api *Api) modemsUpstreamHistory(w http.ResponseWriter, r *http.Request) {
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

    query := db.NewQuery(conn, usHistoryCb, modemUpstreamHistoryQuery)
    query.Where("modem." + column, "=", id)
    query.Where("us.poll_time", ">=", from)
    query.Where("us.poll_time", "<=", to)
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

    addCacheHeader(to, w)
    hist, _ := history.([]*UpstreamHistory)
    w.Header().Set("X-Count", strconv.Itoa(len(hist)))
    JsonResponse(w, hist)
}
