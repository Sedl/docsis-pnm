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

    community := "public"
    for _, cmtso := range api.Manager.GetCmtsList() {
        if cmtso.ValueOfDbId() != mdm.CmtsId {
            continue
        } else {
            community = cmtso.GetModemCommunity(mdm.Mac)
            break
        }
    }

    mdata, err := modem.Poll(mdm.IP.String(), mdm.Mac, community)
    if err != nil {
        HandleServerError(w, err)
        return
    }

    JsonResponse(w, mdata)
}
