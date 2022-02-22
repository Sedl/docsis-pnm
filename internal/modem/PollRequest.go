package modem

import (
    "fmt"
    "github.com/sedl/docsis-pnm/internal/constants"
    snmp2 "github.com/sedl/docsis-pnm/internal/snmp"
    "github.com/sedl/docsis-pnm/internal/types"
    "github.com/gosnmp/gosnmp"
    "log"
    "net"
    "strconv"
    "time"
)

type Poller struct {
    Hostname string // hostname or IP
    // CmtsId The database ID where the modem is currently active. This is here so the database gets an update of the
    // modems location. The location can change because customers tend to move...
    CmtsId    int32
    Mac       net.HardwareAddr
    SnmpIndex int32
    Community string
    snmp *gosnmp.GoSNMP
}

func (p *Poller) Connect() error {
   if p.snmp == nil {
       // TODO: timeout from config
       p.snmp = &gosnmp.GoSNMP{
           Target:         p.Hostname,
           Community:      p.Community,
           Version:        gosnmp.Version2c,
           Timeout:        time.Duration(5) * time.Second,
           Retries:        3,
           Port:           161,
           MaxRepetitions: 10,
           ExponentialTimeout: false,
       }
       return p.snmp.Connect()
   }
   return nil
}

func (p *Poller) Close() error {
    if p.snmp != nil {
        return p.snmp.Conn.Close()
    }
    return nil
}

func (p *Poller) Poll() (data *types.ModemData, err error) {
    start := time.Now()

    data = &types.ModemData{
        Hostname: p.Hostname,
        Mac: p.Mac,
        Timestamp: start.Unix(),
    }

    err = p.modemBasicInfo(data)
    if err != nil {
        return nil, err
    }

    iftypes, err := p.getInterfaceTypes()
    if err != nil {
        return nil, err
    }

    // Look for OFDM downstream and MAC interface
    ofdmDownstreamId := 0
    macInterface := 0
    for idx, value := range iftypes {
        if value == 277 {
            ofdmDownstreamId = idx
        } else if value == 127 {
            macInterface = idx
        }
    }

    if ofdmDownstreamId > 0 {
        ofdms, err := p.getOfdmDownstreams(ofdmDownstreamId)
        if err != nil {
            return nil, err
        }
        data.OfdmDownstreams = ofdms
    }

    if data.DownStreams, data.DSPrimary, err = p.GetDownstreamChannels(); err != nil {
        log.Printf(
            "Error getting downstream information from modem %s: %s", p.Hostname, err)
        data.Err = err
        return
    }

    if data.UpStreams, err = p.GetUpstreamChannels(data.DocsisVersion); err != nil {
        log.Printf(
            "Error getting upstream information from modem %s: %s", p.Hostname, err)
        data.Err = err
        return
    }

    if err = p.getByteCount(macInterface, data); err != nil {
        log.Printf("error getting byte counters of MAC interface from modem %s: %s", p.Hostname, err)
        data.Err = err
        return
    }

    data.QueryTime = int64(time.Now().Sub(start))
    return
}

func (p *Poller) getByteCount(macIntfIndex int, mdata *types.ModemData) error {
    if macIntfIndex == 0 {
        return nil
    }
    idx := fmt.Sprintf(".%d", macIntfIndex)
    result, err := p.snmp.Get([]string{
        constants.IfHCInOctets + idx,
        constants.IfHCOutOctets + idx,
        constants.IfInOctets + idx,
        constants.IfOutOctets + idx,
    })
    if err != nil {
        return err
    }

    var outOctets, inOctets uint32

    for _, pdu := range result.Variables {
        oid, _ := snmp2.SliceOID(pdu.Name)
        switch oid {
        case constants.IfHCOutOctets:
            mdata.BytesUp, _ = snmp2.ToUint64(&pdu)
        case constants.IfHCInOctets:
            mdata.BytesDown, _ = snmp2.ToUint64(&pdu)
        case constants.IfInOctets:
            inOctets,  _ = snmp2.ToUint32(&pdu)
        case constants.IfOutOctets:
            outOctets, _ = snmp2.ToUint32(&pdu)
        }
    }

    // use 32 bit counters as alternative source
    if mdata.BytesUp == 0 && outOctets > 0 {
        mdata.BytesUp = uint64(outOctets)
    }
    if mdata.BytesDown == 0 && inOctets > 0 {
        mdata.BytesDown = uint64(inOctets)
    }

    return nil
}

func (p *Poller) modemBasicInfo(data *types.ModemData) error {
    // We have to get all the values seperately because some modems (like Teleste amplifiers) don't support
    // multiple SNMP get values in one request and will return NULL values

    var docsCap, docs31Cap int
    result, err := p.snmp.Get([]string{constants.SysDescr})
    if err != nil {
        return err
    }
    if len(result.Variables) > 0 {
        data.Sysdescr, err = snmp2.ToString(&result.Variables[0])
    }

    // check for DOCSIS capabilities
    result, err = p.snmp.Get([]string{constants.DocsIf31CmDocsisBaseCapability})
    if err == nil && len(result.Variables) > 0 && result.Variables[0].Type == gosnmp.Integer {
        docs31Cap, _ = snmp2.ToInt(&result.Variables[0])
    }

    if docs31Cap > 0 {
        data.DocsisVersion = constants.DocsVer31
    } else {
        // probably not a DOCSIS 3.1 modem, check for older versions
        result, err = p.snmp.Get([]string{constants.DocsIfDocsisBaseCapability})
        if err == nil && len(result.Variables) > 0 {
            docsCap, _ = snmp2.ToInt(&result.Variables[0])
        }

        if docsCap >= 4 {
            data.DocsisVersion = constants.DocsVer30
        } else {
            data.DocsisVersion = uint32(docsCap)
        }
    }

    // uptime
    result, err = p.snmp.Get([]string{constants.SysUpTimeInstance})
    if err == nil && len(result.Variables) > 0 {
        uptime, _ := snmp2.ToUint32(&result.Variables[0])
        data.Uptime = uptime
    }

    return nil
}


func (p *Poller) getInterfaceTypes() (map[int]int, error) {
    result, err := p.snmp.BulkWalkAll(constants.IfType)
    if err != nil {
        return nil, err
    }

    iftypes := make(map[int]int)
    for _, ift := range result {
        _, idx := snmp2.SliceOID(ift.Name)
        iftypes[idx], err = snmp2.ToInt(&ift)
        if err != nil {
            return nil, err
        }
    }

    return iftypes, nil
}

func (p *Poller) GetUpstreamChannels(docsisVersion uint32) ([]types.UpstreamChannel, error) {
    results, err := p.snmp.BulkWalkAll(".1.3.6.1.2.1.10.127.1.1.2.1")
    if err != nil {
        return nil, err
    }

    upstreams := make(map[int]*types.UpstreamChannel)

    var upstr *types.UpstreamChannel
    var ok bool

    for _, result := range results {
        oid, idx := snmp2.SliceOID(result.Name)

        if upstr, ok = upstreams[idx]; !ok {
            upstr = &types.UpstreamChannel{Index: int32(idx)}
            upstreams[idx] = upstr
        }

        switch oid {
        case constants.DocsIfUpChannelId:
            upstr.ID, _ = snmp2.ToInt32(&result)

        case constants.DocsIfUpChannelFrequency:
            upstr.Freq, _ = snmp2.ToInt32(&result)

        case constants.DocsIfUpChannelWidth:
            upstr.Width, _ = snmp2.ToInt32(&result)

        case constants.DocsIfUpChannelTxTimingOffset:
            upstr.TimingOffset, _ = snmp2.ToUint32(&result)

        case constants.DocsIf3CmStatusUsTxPower:
            upstr.TxPower, _ = snmp2.ToInt32(&result)
        }
    }

    if docsisVersion >= constants.DocsVer30 {
        // Get additional upstream metrics from DOCS-IF3 subtree
        results, err = p.snmp.BulkWalkAll(constants.DocsIf3CmStatusUsTxPower)

        for _, result := range results {

            _, idx := snmp2.SliceOID(result.Name)

            if upstr, ok = upstreams[idx]; !ok {
                continue
            }

            upstr.TxPower, _ = snmp2.ToInt32(&result)
        }
    } else if len(upstreams) > 0 {
        // Before DOCSIS 3.0 we have only one upstream channel
        results, err = p.snmp.BulkWalkAll(constants.DocsIfCmStatusTxPower)
        if err != nil {
            return nil, err
        }
        for _, v := range upstreams {
            v.TxPower, _ = snmp2.ToInt32(&results[0])
        }
        if len(results) > 0 {
            for _, v := range upstreams {
                v.TxPower, _ = snmp2.ToInt32(&results[0])
                break
            }
        }
    }

    uslist := make([]types.UpstreamChannel, 0)
    for _, ds := range upstreams {
        uslist = append(uslist, *ds)
    }
    return uslist, nil

}

func (p *Poller) getOfdmDownstreams(downstreamIndex int)  (*types.OfdmDownstreamHistory, error) {

    oids := []string{
        constants.DocsIf31CmDsOfdmChannelPowerCenterFrequency,
        constants.DocsIf31CmDsOfdmChannelPowerRxPower,
    }

    appendix := "." + strconv.Itoa(downstreamIndex)
    subtree, err := snmp2.WalkSubtree(p.snmp, oids)
    if err != nil {
        return nil, err
    }

    oid := constants.DocsIf31CmDsOfdmChannelPowerCenterFrequency + appendix
    pduMap := subtree[oid]
    if pduMap == nil {
        // no OFDM downstream for this downstreamIndex available
        return nil, nil
    }

    downstreams := make(map[int]*types.OfdmDownstream)
    for idx, pdu := range pduMap {
        freq, err := snmp2.ToUint32(pdu)
        if err != nil {
            return nil, err
        }
        downstreams[idx] = &types.OfdmDownstream{
            Freq: freq,
        }
    }

    oid = constants.DocsIf31CmDsOfdmChannelPowerRxPower + appendix
    pduMap = subtree[oid]
    for idx, pdu := range pduMap {
        power, err := snmp2.ToInt(pdu)
        if err != nil {
            return nil, err
        }
        if downstr, ok := downstreams[idx]; ok {
            downstr.Power = float32(power) / 10
        }
    }

    history := &types.OfdmDownstreamHistory{
        Downstreams: downstreams,
    }

    // collect profile information

    oids = []string{
        // this is the whole profile subtree
        ".1.3.6.1.4.1.4491.2.1.28.1.10.1",
    }
    subtree, err = snmp2.WalkSubtree(p.snmp, oids)
    if err != nil {
        return nil, err
    }

    profiles := make(map[int]*types.OfdmDownstreamChannelProfileData)

    oid = constants.DocsIf31CmDsOfdmProfileStatsConfigChangeCt + appendix
    for idx, pdu := range subtree[oid] {
        chCount, err := snmp2.ToUint32(pdu)
        if err != nil {
            return nil, err
        }
        profiles[idx] = &types.OfdmDownstreamChannelProfileData{
            ChangeCount:      chCount,
        }
    }

    oid = constants.DocsIf31CmDsOfdmProfileStatsTotalCodewords + appendix
    for idx, pdu := range subtree[oid] {
        if pdu.Type == gosnmp.Null {
            continue
        }
        cwTotal, err := snmp2.ToUint64(pdu)
        if err != nil {
            return nil, err
        }
        prof := profiles[idx]
        if prof != nil {
            prof.CwTotal = cwTotal
        }
    }

    oid = constants.DocsIf31CmDsOfdmProfileStatsCorrectedCodewords + appendix
    for idx, pdu := range subtree[oid] {
        if pdu.Type == gosnmp.Null {
            continue
        }
        cwCorr, err := snmp2.ToUint64(pdu)
        if err != nil {
            return nil, err
        }
        prof := profiles[idx]
        if prof != nil {
            prof.CwCorrected = cwCorr
        }
    }

    oid = constants.DocsIf31CmDsOfdmProfileStatsUncorrectableCodewords + appendix
    for idx, pdu := range subtree[oid] {
        if pdu.Type == gosnmp.Null {
            continue
        }
        cwUncorr, err := snmp2.ToUint64(pdu)
        if err != nil {
            return nil, err
        }
        prof := profiles[idx]
        if prof != nil {
            prof.CwUncorrectables = cwUncorr
        }
    }

    history.ProfileData = profiles
    return history, nil
}

func (p *Poller) GetDownstreamChannels() ([]types.DownstreamChannel, int32, error) {

    // docsIftypes.DownstreamChannelEntry
    res1, err := p.snmp.BulkWalkAll(".1.3.6.1.2.1.10.127.1.1.1.1")
    if err != nil {
        return nil, 0, err
    }

    // docsIfSignalQualityEntry
    res2, err := p.snmp.BulkWalkAll(".1.3.6.1.2.1.10.127.1.1.4.1")
    if err != nil {
        return nil, 0, err
    }

    results := append(res1, res2...)

    downstreams := make(map[int]*types.DownstreamChannel)

    var downstr *types.DownstreamChannel
    var ok bool
    var primary int32

    for _, result := range results {
        oid, idx := snmp2.SliceOID(result.Name)

        if downstr, ok = downstreams[idx]; !ok {
            downstr = &types.DownstreamChannel{Index: idx}
            downstreams[idx] = downstr
        }

        switch oid {

        // DOCS-IF-MIB::docsIfDownChannelId
        case ".1.3.6.1.2.1.10.127.1.1.1.1.1":
            downstr.ID, _ = snmp2.ToInt32(&result)

        // DOCS-IF-MIB::docsIfDownChannelFrequency
        case ".1.3.6.1.2.1.10.127.1.1.1.1.2":
            downstr.Freq, _ = snmp2.ToInt32(&result)
            if idx == 3 {
                primary = downstr.Freq
            }

        // DOCS-IF-MIB::docsIfDownChannelPower
        case ".1.3.6.1.2.1.10.127.1.1.1.1.6":
            if pow, err1 := snmp2.ToInt32(&result); err1 == nil {
                downstr.Power = pow
            }

        // DOCS-IF-MIB::docsIfSigQSignalNoise
        case ".1.3.6.1.2.1.10.127.1.1.4.1.5":
            if snr, err1 := snmp2.ToInt32(&result); err1 == nil {
                downstr.SNR = snr
            }

        // DOCS-IF-MIB::docsIfSigQMicroreflections
        case ".1.3.6.1.2.1.10.127.1.1.4.1.6":
            if refl, err1 := snmp2.ToInt32(&result); err1 == nil {
                downstr.Microrefl = refl * -1
            }

        // DOCS-IF-MIB::docsIfSigQExtUnerroreds
        case ".1.3.6.1.2.1.10.127.1.1.4.1.8":
            downstr.Unerroreds, _ = snmp2.ToUint64(&result)

        // DOCS-IF-MIB::docsIfSigQExtCorrecteds
        case ".1.3.6.1.2.1.10.127.1.1.4.1.9":
            downstr.Correcteds, _ = snmp2.ToUint64(&result)

        // DOCS-IF-MIB::docsIfSigQExtUncorrectables
        case ".1.3.6.1.2.1.10.127.1.1.4.1.10":
            downstr.Uncorrectables, _ = snmp2.ToUint64(&result)

        // DOCS-IF-MIB::docsIfDownChannelModulation
        case ".1.3.6.1.2.1.10.127.1.1.1.1.4":
            downstr.Modulation, _ = snmp2.ToInt32(&result)
        }
    }


    dslist := make([]types.DownstreamChannel, 0)
    for _, ds := range downstreams {
        dslist = append(dslist, *ds)
    }
    return dslist, primary, nil
}