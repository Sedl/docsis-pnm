package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/sedl/docsis-pnm/internal/types"
	"net/http"
	"strconv"
)

func (api *Api) cmtsList(w http.ResponseWriter, r *http.Request) {
	db := api.Manager.GetDbInterface()
	cmtsList , err := db.GetCMTSAll()
	if err != nil {
		HandleServerError(w, err)
		return
	}
	JsonResponse(w, cmtsList)
}

func (api *Api) cmtsOne(w http.ResponseWriter, r *http.Request) {
	db := api.Manager.GetDbInterface()

	id := mux.Vars(r)["cmtsId"]
	iid, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		HandleServerError(w, err)
		return
	}

	cmts , err := db.GetCMTSById(uint32(iid))
	if err != nil {
		HandleServerError(w, err)
		return
	}

	if cmts == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	JsonResponse(w, cmts)
}

func IsDuplicateError(err error) bool {
	if perr, ok := err.(*pq.Error); ok {
		if perr.Code == "23505" {
			return true
		}
	}
	return false
}

func (api *Api) cmtsCreate(w http.ResponseWriter, r* http.Request) {
	var cmts types.CMTSRecord
	err := json.NewDecoder(r.Body).Decode(&cmts)
	if err != nil {
		HandleServerError(w, err)
		return
	}

	err = api.Manager.GetDbInterface().InsertCMTS(&cmts)
	if err != nil {
		if IsDuplicateError(err) {
			HandleServerConflict(w, "An entry with this hostname already exists")
		} else {
			HandleServerError(w, err)
		}
		return
	}

	if cmts.Disabled == false {
		// start stuff
		cmtsobj, err := api.Manager.AddCMTS(&cmts)
		if err != nil {
			HandleServerError(w, err)
			return
		}
		err = cmtsobj.Run()
		if err != nil {
			HandleServerError(w, err)
			return
		}
	}
	JsonResponse(w, &cmts)
	fmt.Printf("%#v", cmts)
}