package api

import (
    "github.com/gorilla/mux"
    "net/http"
    "regexp"
    "strconv"
)

type ParsedPath struct {
    ModemColumn string
    ModemId interface{}
    FromTs int64
    ToTs int64
}

func detectModemIdUrlColumn(modemId string) (string, error) {
    matched , err := regexp.MatchString("^[0-9]+$", modemId)
    if err != nil {
        return "", err
    }
    if matched {
        return "id", nil
    }

    matched, err = regexp.MatchString("^[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+$", modemId)
    if err != nil {
        return "", err
    }
    if matched {
        return "ip", nil
    }

    matched, err = regexp.MatchString("^[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}$", modemId)
    if err != nil {
        return "", err
    }
    if matched {
        return "mac", nil
    }

    return "", ErrorInvalidModemId
}

func ParsePath(r *http.Request) (ParsedPath, error) {
    var err error
    ppath := ParsedPath{}

    vars := mux.Vars(r)
    if val, ok := vars["modemId"]; ok {
        ppath.ModemColumn, err = detectModemIdUrlColumn(val)
        ppath.ModemId = val
        if err != nil {
            return ppath, err
        }
    }

   if val, ok := vars["fromTS"] ; ok {
       ppath.FromTs, err = strconv.ParseInt(val, 10, 64)
       if err != nil {
           return ppath, err
       }
   }

    if val, ok := vars["toTS"] ; ok {
        ppath.ToTs, err = strconv.ParseInt(val, 10, 64)
        if err != nil {
            return ppath, err
        }
    }

    return ppath, nil
}