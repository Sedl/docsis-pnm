package api

import (
    "database/sql"
    "github.com/sedl/docsis-pnm/internal/db"
    "net/http"
)

const modemUpstreamHistoryQuery = `
SELECT
    u.modem_id,
    u.poll_time,
    u.freq,
    u.timing_offset,
    u.tx_power
FROM
    modem_upstream AS u
JOIN
    modem
        ON (
            modem.id = u.modem_id
        )`

type ModemUpstreamRecord struct {
    ModemID           int32  `json:"-"`
    PollTime 		  int32  `json:"-"`
    Freq              int32  `json:"freq"`
    TimingOffset      uint32 `json:"timing_offset"`
    TxPower           int32  `json:"tx_power"`
}

type ModemUpstreamHistory struct {
    TS int64 `json:"ts"`
    US []*ModemUpstreamRecord `json:"us"`
}

func usModemHistoryCb(rows *sql.Rows) (interface{}, error) {

    history := make([]*ModemUpstreamHistory, 0)

    var lastTS, ts int64
    var err error

    var hist *ModemUpstreamHistory
    for rows.Next() {

        us := &ModemUpstreamRecord{}
        err = rows.Scan(
            &us.ModemID,
            &ts,
            &us.Freq,
            &us.TimingOffset,
            &us.TxPower,
        )
        if err != nil {
            return nil, err
        }

        if hist == nil || lastTS != ts {
            hist = &ModemUpstreamHistory{
                TS: ts,
                US: make([]*ModemUpstreamRecord, 0),
            }
            history = append(history, hist)
            lastTS = ts
        }
        hist.US = append(hist.US, us)

    }
    return history, nil
}

func (api *Api) modemUpstream (w http.ResponseWriter, r *http.Request) {
    pvars, err := ParsePath(r)
    if err != nil {
        HandleServerError(w, err)
        return
    }

    conn, err := api.Manager.GetDbInterface().GetConn()
    if err != nil {
        HandleServerError(w, err)
        return
    }


    query := db.NewQuery(conn, usModemHistoryCb, modemUpstreamHistoryQuery)
    query.Where("modem." + pvars.ModemColumn, "=", pvars.ModemId)
    query.Where("u.poll_time", ">=", pvars.FromTs)
    query.Where("u.poll_time", "<=", pvars.ToTs)

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

    hist := history.([]*ModemUpstreamHistory)

    addCacheHeader(pvars.ToTs, w)
    addCountHeader(w, len(hist))
    JsonResponse(w, hist)
}

func (api *Api) modemUpstreamLatest (w http.ResponseWriter, r *http.Request) {
    pvars, err := ParsePath(r)
    if err != nil {
        HandleServerError(w, err)
        return
    }

    conn, err := api.Manager.GetDbInterface().GetConn()
    if err != nil {
        HandleServerError(w, err)
        return
    }

    query := db.NewQuery(conn, usModemHistoryCb, modemUpstreamHistoryQuery)
    query.Where("modem." + pvars.ModemColumn, "=", pvars.ModemId)
    query.Where("u.poll_time", "=", db.SqlExpression("(SELECT MAX(poll_time) FROM modem_upstream WHERE modem_id = modem.id)"))
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

    hist := history.([]*ModemUpstreamHistory)
    if len(hist) == 0 {
        JsonResponse(w, nil)
        return
    }

    JsonResponse(w, hist[0])
}
