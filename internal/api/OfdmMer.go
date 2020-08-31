package api

import (
    "errors"
    db2 "github.com/sedl/docsis-pnm/internal/db"
    "github.com/sedl/docsis-pnm/internal/misc"
    "github.com/sedl/docsis-pnm/internal/modem"
    "github.com/sedl/docsis-pnm/internal/parse"
    tftp2 "github.com/sedl/docsis-pnm/internal/tftp"
    "log"
    "net/http"
    "time"
)

func (api *Api) OfdmMer (w http.ResponseWriter, r *http.Request) {
    tftp := api.Manager.GetTftpServerInstance()
    if tftp == nil {
        HandleServerError(w, errors.New("TFTP functionality needed but not enabled. Please check the tftp.externaladdress config option"))
        return
    }

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


    filename := misc.RandomFilename(16)
    log.Println(filename)
    fileChan := make(chan *tftp2.SafeBuffer)
    // prepare for file retrieval
    go func() {
        file, errf := tftp.WaitForFile(filename, time.Second*120)
        if errf != nil {
            log.Printf("error while receiving file via TFTP for modem %s (%s): %s\n", mdm.Mac, mdm.IP, errf )
            fileChan<- nil
        } else {
            fileChan<- file
        }
    }()

    request := &modem.Poller{
        Hostname:  mdm.IP.String(),
        Community: api.Manager.GetCmtsModemCommunity(mdm.CmtsId),
    }

    err = request.Connect()
    if err != nil {
        HandleServerError(w, err)
        return
    }
    defer misc.CloseOrLog(request)

    // tftpIp := net.ParseIP("37.156.95.73").To4()
    err = request.RequestPnmOfdmMerFile(filename, tftp.ExternalIp)
    if err != nil {
        log.Println(err)
        HandleServerError(w, err)
        return
    }

    // receive file
    merFile := <-fileChan
    if merFile == nil {
        HandleServerError(w, errors.New("timeout while receiving file"))
        return
    }

    file, err := parse.OfdmMerFile(merFile.Bytes())
    if err != nil {
        HandleServerError(w, err)
        return
    }

    JsonResponse(w, file)
}