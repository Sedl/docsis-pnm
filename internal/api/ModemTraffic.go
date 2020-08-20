package api

import (
    "database/sql"
    "github.com/sedl/docsis-pnm/internal/db"
    "github.com/sedl/docsis-pnm/internal/misc"
    "net/http"
)

const modemTrafficQuery = `
SELECT
    m.poll_time,
    m.bytes_down,
    m.bytes_up
FROM
    modem_data AS m
JOIN
    modem
        ON (
            modem.id = m.modem_id
        )
`

func (api *Api)ModemTraffic(w http.ResponseWriter, r *http.Request) {
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

    query := db.NewQuery(conn, modemTrafficCb, modemTrafficQuery)
    query.Where("m.error_timeout", "=", false)
    query.Where("modem." + vars.ModemColumn, "=", vars.ModemId)
    query.Where("m.poll_time", ">=", vars.FromTs)
    query.Where("m.poll_time", "<=", vars.ToTs)
    query.OrderBy("m.poll_time DESC")
    defer misc.CloseOrLog(query)
    err = query.Exec()
    if err != nil {
        HandleServerError(w, err)
        return
    }

    values := make([][3]uint64, 0)

    for {
        value, err := query.Next()
        if value == nil {
            break
        }
        if err != nil {
            HandleServerError(w, err)
            return
        }
        values = append(values, value.([3]uint64))
    }

    addCacheHeader(vars.ToTs, w)
    addCountHeader(w, len(values))

    JsonResponse(w, values)
}

func modemTrafficCb(rows *sql.Rows) (interface{}, error) {

    if ! rows.Next() {
        return nil, nil
    }

    var values [3]uint64

    err := rows.Scan(&values[0], &values[1], &values[2])
    if err != nil {
        return nil, err
    }

    return values, nil
}