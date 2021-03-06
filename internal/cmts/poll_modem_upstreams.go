package cmts

import (
    "github.com/sedl/docsis-pnm/internal/constants"
    "github.com/sedl/docsis-pnm/internal/snmp"
    "github.com/sedl/docsis-pnm/internal/types"
    "log"
    "time"
)

func (cmts *Cmts) pollModemUpstreams() (error, map[int]map[int]*types.UpstreamModemCMTS) {

    pollTime := time.Now()
    var upstream *types.UpstreamModemCMTS
    modems := make(map[int]map[int]*types.UpstreamModemCMTS)

    getUpstreamRecord := func (usIdx, mdmIdx int) *types.UpstreamModemCMTS {
        var mdm map[int]*types.UpstreamModemCMTS
        var us *types.UpstreamModemCMTS

        mdm, ok := modems[mdmIdx]
        if ! ok {
            modems[mdmIdx] = make(map[int]*types.UpstreamModemCMTS)
            mdm = modems[mdmIdx]
        }
        var usId int32
        var mdmId types.ModemId

        //cmts.upstreamCache.GetByIndex(usIdx)
        us, ok = mdm[usIdx]
        if ! ok {
            usrec := cmts.upstreamCache.GetByIndex(usIdx)
            if usrec == nil {
                log.Printf("error: can't find upstream with id %d in cache (%s)", usIdx, cmts.ValueOfHostname())
                usId = 0
            } else {
                usId = usrec.ID
            }
            mdmRec := cmts.modemList.ByIndex(int32(mdmIdx))
            if mdmRec == nil {
                mdmId = 0
            } else {
                mdmId = mdmRec.DbId
            }
            mdm[usIdx] = &types.UpstreamModemCMTS{
                ModemId: mdmId,
                UpstreamId:     usId,
                PollTime:       int32(pollTime.Unix()),
            }
            us = mdm[usIdx]
        }
        return us
    }

    results, err := cmts.Snmp.BulkWalkAll(constants.DocsIf3CmtsCmUsStatusRxPower)
    if err != nil {
        return err, nil
    }
    for _, result := range results {
        oid, usid := snmp.SliceOID(result.Name)
        oid, mdmidx := snmp.SliceOID(oid)

        power, err := snmp.ToInt32(&result)
        if err != nil {
            log.Printf("warning: error while fetching %s from %s: %s)\n", constants.DocsIf3CmtsCmUsStatusRxPower, cmts.ValueOfHostname(), err.Error())
        }
        upstream = getUpstreamRecord(usid, mdmidx)
        upstream.PowerRx = power
    }

    results, err = cmts.Snmp.BulkWalkAll(constants.DocsIf3CmtsCmUsStatusSignalNoise)
    if err != nil {
        return err, nil
    }
    for _, result := range results {
        oid, usid := snmp.SliceOID(result.Name)
        oid, mdmidx := snmp.SliceOID(oid)

        snr, err := snmp.ToInt32(&result)
        if err != nil {
            log.Printf("warning: error while fetching %s from %s: %s)\n", constants.DocsIf3CmtsCmUsStatusSignalNoise, cmts.ValueOfHostname(), err.Error())
        }
        upstream = getUpstreamRecord(usid, mdmidx)
        upstream.SNR = snr
    }

    results, err = cmts.Snmp.BulkWalkAll(constants.DocsIf3CmtsCmUsStatusMicroreflections)
    if err != nil {
        return err, nil
    }
    for _, result := range results {
        oid, usid := snmp.SliceOID(result.Name)
        oid, mdmidx := snmp.SliceOID(oid)

        microrefl, err  := snmp.ToInt32(&result)
        if err != nil {
            log.Printf("warning: error while fetching %s from %s: %s)\n", constants.DocsIf3CmtsCmUsStatusMicroreflections, cmts.ValueOfHostname(), err.Error())
        }
        upstream = getUpstreamRecord(usid, mdmidx)
        upstream.Microrefl = microrefl
    }

    results, err = cmts.Snmp.BulkWalkAll(constants.DocsIf3CmtsCmUsStatusUnerroreds)
    if err != nil {
        return err, nil
    }
    for _, result := range results {
        oid, usid := snmp.SliceOID(result.Name)
        oid, mdmidx := snmp.SliceOID(oid)

        unerroreds, err  := snmp.ToInt64(&result)
        if err != nil {
            log.Printf("warning: error while fetching %s from %s: %s)\n", constants.DocsIf3CmtsCmUsStatusUnerroreds, cmts.ValueOfHostname(), err.Error())
        }
        upstream = getUpstreamRecord(usid, mdmidx)
        upstream.Unerroreds = unerroreds
    }

    results, err = cmts.Snmp.BulkWalkAll(constants.DocsIf3CmtsCmUsStatusCorrecteds)
    if err != nil {
        return err, nil
    }
    for _, result := range results {
        oid, usid := snmp.SliceOID(result.Name)
        oid, mdmidx := snmp.SliceOID(oid)

        correcteds, err  := snmp.ToInt64(&result)
        if err != nil {
            log.Printf("warning: error while fetching %s from %s: %s)\n", constants.DocsIf3CmtsCmUsStatusCorrecteds, cmts.ValueOfHostname(), err.Error())
        }
        upstream = getUpstreamRecord(usid, mdmidx)
        upstream.Correcteds = correcteds
    }

    results, err = cmts.Snmp.BulkWalkAll(constants.DocsIf3CmtsCmUsStatusUncorrectables)
    if err != nil {
        return err, nil
    }
    for _, result := range results {
        oid, usid := snmp.SliceOID(result.Name)
        oid, mdmidx := snmp.SliceOID(oid)

        uncorr, err  := snmp.ToInt64(&result)
        if err != nil {
            log.Printf("warning: error while fetching %s from %s: %s)\n", constants.DocsIf3CmtsCmUsStatusUncorrectables, cmts.ValueOfHostname(), err.Error())
        }
        upstream = getUpstreamRecord(usid, mdmidx)
        upstream.Erroreds = uncorr
    }

    return nil, modems
}
