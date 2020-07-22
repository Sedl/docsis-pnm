package api

import (
	"github.com/gorilla/mux"
	db2 "github.com/sedl/docsis-pnm/internal/db"
	"github.com/sedl/docsis-pnm/internal/modem"
	"github.com/sedl/docsis-pnm/internal/types"
	"net"
	"net/http"
	"strconv"
)

type ModemJson struct {
	Id            uint64 `json:"modem_id"`
	CmtsId        uint32 `json:"cmts_id"`
	Mac           string `json:"mac"`
	Sysdescr      string `json:"sysdescr"`
	IPAddr        string `json:"ipaddr"`
	SnmpIndex     uint32 `json:"snmp_index"`
	DocsisVersion string `json:"docsis_version"`
}

func convertToModemJson(record *types.ModemRecord) *ModemJson {

	var docsver string
	switch record.DocsisVersion {
	case modem.DocsVers10:
		docsver = "docs_10"
	case modem.DocsVer11:
		docsver = "docs_11"
	case modem.DocsVer20:
		docsver = "docs_20"
	case modem.DocsVer30:
		docsver = "docs_30"
	case modem.DocsVer31:
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

func (api *Api) modemsAll(w http.ResponseWriter, r *http.Request) {
	api.modemsBy(w, "", false)
}

func (api *Api) modemsById(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["modemId"]
	mid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		HandleServerError(w, err)
		return
	}

	api.modemsBy(w, "id = $1", true, mid)
}

func (api *Api) modemsByMac(w http.ResponseWriter, r *http.Request) {
	mac := mux.Vars(r)["modemMac"]
	hwaddr, err := net.ParseMAC(mac)
	if err != nil {
		// invalid MAC address
		// we don't care, just return a 404
		w.WriteHeader(http.StatusNotFound)
		return
	}

	api.modemsBy(w, "mac = $1", true, hwaddr.String())
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

func (api *Api) modemsByIp(w http.ResponseWriter, r* http.Request) {
	ipstr := mux.Vars(r)["modemIp"]
	ip := net.ParseIP(ipstr)

	if ip == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	api.modemsBy(w, "ip = $1", true, ip.String())

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
	defer db2.CloseOrLog(query)

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
