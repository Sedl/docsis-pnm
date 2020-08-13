package api

import (
	"github.com/gorilla/mux"
	consts "github.com/sedl/docsis-pnm/internal/constants"
	db2 "github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/misc"
	"github.com/sedl/docsis-pnm/internal/types"
	"net/http"
	"strconv"
)

type ModemJson struct {
	Id            uint64 `json:"modem_id"`
	CmtsId        int32  `json:"cmts_id"`
	Mac           string `json:"mac"`
	Sysdescr      string `json:"sysdescr"`
	IPAddr        string `json:"ipaddr"`
	SnmpIndex     uint32 `json:"snmp_index"`
	DocsisVersion string `json:"docsis_version"`
}

func convertToModemJson(record *types.ModemRecord) *ModemJson {

	var docsver string
	switch record.DocsisVersion {
	case consts.DocsVers10:
		docsver = "docs_10"
	case consts.DocsVer11:
		docsver = "docs_11"
	case consts.DocsVer20:
		docsver = "docs_20"
	case consts.DocsVer30:
		docsver = "docs_30"
	case consts.DocsVer31:
		docsver = "docs_31"
	default:
		docsver = "unknown"
	}

	js := &ModemJson{
		Id:            record.ID,
		CmtsId:        record.CmtsId,
		Mac:           record.Mac.String(),
		Sysdescr:      record.SysDescr,
		IPAddr:        record.IP.String(),
		SnmpIndex:     uint32(record.SnmpIndex),
		DocsisVersion: docsver,
	}

	return js
}

func (api *Api) modemsAll(w http.ResponseWriter, _ *http.Request) {
	api.modemsBy(w, "", false)
}

func (api *Api) modemsById(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["modemId"]
	where, err := detectModemIdUrlColumn(id)
	if err != nil {
		HandleBadRequest(w, err)
		return
	}
	api.modemsBy(w, where+" = $1", true, id)
}

func (api *Api) modemsByCmtsId(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["cmtsId"]
	cmtsId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		HandleServerError(w, err)
		return
	}

	api.modemsBy(w, "cmts_id = $1", false, cmtsId)
}

func (api *Api) modemsBy(w http.ResponseWriter, where string, single bool, args ...interface{}) {

	conn, err := api.Manager.GetDbInterface().GetConn()
	if err != nil {
		HandleServerError(w, err)
		return
	}

	query, err := db2.NewModemQuery(conn, where, args...)
	if err != nil {
		HandleServerError(w, err)
		return
	}
	defer misc.CloseOrLog(query)

	if single {
		mdm, err := query.Next()
		if err != nil {
			HandleServerError(w, err)
			return
		}
		if mdm == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		JsonResponse(w, convertToModemJson(mdm))

	} else {

		modems := make([]*ModemJson, 0)
		for {
			mdm, err := query.Next()
			if err != nil {
				HandleServerError(w, err)
				return
			}
			if mdm == nil {
				break
			}
			modems = append(modems, convertToModemJson(mdm))
		}
		JsonResponse(w, modems)
	}
}
