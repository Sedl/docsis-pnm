package api

import (
    "database/sql"
    "github.com/gorilla/mux"
    "github.com/sedl/docsis-pnm/internal/db"
    "github.com/sedl/docsis-pnm/internal/types"
    "net/http"
)

const cmtsUpstreamQuery = `SELECT id, cmts_id, snmp_idx, descr, freq, alias, admin_status FROM cmts_upstream`

func upstreamCb(rows *sql.Rows) (interface{}, error) {
    if ! rows.Next() {
        return nil, nil
    }
    us := types.CMTSUpstreamRecord{}

    err := rows.Scan(&us.ID, &us.CMTSID, &us.SNMPIndex, &us.Description, &us.Freq, &us.Alias, &us.AdminStatus)
    if err != nil {
        return nil, err
    }

    return us, nil
}

func (api *Api) upstreamsById(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["upstreamId"]

    conn, err := api.Manager.GetDbInterface().GetConn()
    if err != nil {
        HandleServerError(w, err)
        return
    }
    query := db.NewQuery(conn, upstreamCb, cmtsUpstreamQuery)
    query.Where("id", "=", id)
    err = query.Exec()
    if err != nil {
        HandleServerError(w, err)
        return
    }

    rec, err := query.Next()
    if err != nil {
        HandleServerError(w, err)
        return
    }
    if rec == nil {
        w.WriteHeader(404)
        return
    }

    JsonResponse(w, rec)
}

func (api *Api) upstreams(w http.ResponseWriter, _ *http.Request) {

    conn, err := api.Manager.GetDbInterface().GetConn()
    if err != nil {
        HandleServerError(w, err)
        return
    }

    query := db.NewQuery(conn, upstreamCb, cmtsUpstreamQuery)
    err = query.Exec()
    if err != nil {
        HandleServerError(w, err)
        return
    }

    upstreams := make([]*types.CMTSUpstreamRecord, 0)

    for {
        upstr, err := query.Next()
        if err != nil {
            HandleServerError(w, err)
            return
        }
        if upstr == nil {
            break
        }
        us := upstr.(types.CMTSUpstreamRecord)
        upstreams = append(upstreams, &us)
    }

    addCountHeader(w, len(upstreams))
    JsonResponse(w, upstreams)

}