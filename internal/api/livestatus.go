package api

import (
    db2 "github.com/sedl/docsis-pnm/internal/db"
    "github.com/sedl/docsis-pnm/internal/modem"
    "net/http"
)

func (api *Api) modemLiveStatus(w http.ResponseWriter, r *http.Request) {
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

    query, err := db2.NewModemQuery(conn, vars.ModemColumn + " = $1", vars.ModemId)
    if err != nil {
        HandleServerError(w, err)
        return
    }

    mdm, err := query.Next()
    if err != nil {
        HandleServerError(w, err)
        return
    }
    if mdm == nil {
        w.WriteHeader(404)
        return
    }

    community := api.Manager.GetCmtsModemCommunity(mdm.CmtsId)

    if community == "" {
        community = "public"
    }

    poller := modem.Poller{
        Hostname:  mdm.IP.String(),
        Mac:       mdm.Mac,
        SnmpIndex: 0,
        Community: community,
    }

    err = poller.Connect()
    if err != nil {
        HandleServerError(w, err)
        return
    }

    mdata, err := poller.Poll()
    if err != nil {
        HandleServerError(w, err)
        return
    }

    JsonResponse(w, mdata)
}
