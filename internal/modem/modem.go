package modem

import (
	snmp2 "github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/soniah/gosnmp"
	"log"
	"net"
	"time"
)

const (
	docsIfDocsisBaseCapability     = ".1.3.6.1.2.1.10.127.1.1.5.0"
	sysDescr                       = ".1.3.6.1.2.1.1.1.0"
	docsIf31CmDocsisBaseCapability = ".1.3.6.1.4.1.4491.2.1.28.1.1.0"
	DocsVers10                     = 1
	DocsVer11                      = 2
	DocsVer20                      = 3
	DocsVer30                      = 4
	DocsVer31                      = 5
	ifType                         = ".1.3.6.1.2.1.2.2.1.3"
)

func docsisVersion1(packet *gosnmp.SnmpPacket) uint32 {
	var ver uint32
	var err error
	if packet.Variables[2].Type == gosnmp.Integer {
		ver, err = snmp2.ToUint32(&packet.Variables[2])
		if err != nil {
			return 0
		}
		if ver > 0 {
			return DocsVer31
		}
	}

	if packet.Variables[1].Type == gosnmp.Integer {
		ver, err = snmp2.ToUint32(&packet.Variables[1])
		if err != nil {
			return 0
		}
		// some modems return 5 here but don't have DOCSIS 3.1 MIBs
		if ver >= 4 {
			return DocsVer30
		} else {
			return ver
		}
	}

	return 0
}

func modemBasicInfo(snmp *gosnmp.GoSNMP, data *types.ModemData) error {
    // We have to get all the values seperately because some modems (like Teleste amplifiers) don't support
	// multiple SNMP get values in one request and will return NULL values

    var docsCap, docs31Cap int
	result, err := snmp.Get([]string{sysDescr})
	if err != nil {
		return err
	}
	if len(result.Variables) > 0 {
		data.Sysdescr, err = snmp2.ToString(&result.Variables[0])
	}

	result, err = snmp.Get([]string{docsIf31CmDocsisBaseCapability})
	if err == nil && len(result.Variables) > 0 && result.Variables[0].Type == gosnmp.Integer {
			docs31Cap, _ = snmp2.ToInt(&result.Variables[0])
	}

	if docs31Cap > 0 {
		data.DocsisVersion = DocsVer31
	} else {
		// probably no DOCSIS 3.1 modem, check for older versions
		result, err = snmp.Get([]string{docsIfDocsisBaseCapability})
		if err == nil && len(result.Variables) > 0 {
			docsCap, _ = snmp2.ToInt(&result.Variables[0])
		}

		if docsCap >= 4 {
			data.DocsisVersion = DocsVer30
		} else {
			data.DocsisVersion = uint32(docsCap)
		}
	}

	return nil
	/*
	var result *gosnmp.SnmpPacket
	oids := []string{sysDescr, docsIfDocsisBaseCapability, docsIf31CmDocsisBaseCapability}

	result, err := snmp.Get(oids)
	if err != nil {
		return err
	}

	if len(result.Variables) == len(oids) {
		descr, err := snmp2.ToString(&result.Variables[0])
		if err == nil {
			data.Sysdescr = descr
		}
		data.DocsisVersion = docsisVersion(result)
	}
	return nil
	 */
}

func getInterfaceTypes(snmp *gosnmp.GoSNMP) (map[int]int, error) {
	result, err := snmp.BulkWalkAll(ifType)
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

// Poll collects performance data from a modem. The mac is optional, it is for logging purposes only
func Poll(ip string, mac net.HardwareAddr, community string) (data *types.ModemData, err error) {
	start := time.Now()

	var macs string

	// TODO: timeout from config
	snmp := &gosnmp.GoSNMP{
		Target:         ip,
		Community:      community,
		Version:        gosnmp.Version2c,
		Timeout:        time.Duration(5) * time.Second,
		Retries:        3,
		Port:           161,
		MaxRepetitions: 10,
		ExponentialTimeout: false,
	}

	data = &types.ModemData{
		Hostname: ip,
		Mac: mac,
		Timestamp: start.Unix(),
	}

	if err = snmp.Connect(); err != nil {
		return
	}

	defer func () {
		err := snmp.Conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()


	err = modemBasicInfo(snmp, data)
	if err != nil {
		return nil, err
	}

	iftypes, err := getInterfaceTypes(snmp)
	if err != nil {
		return nil, err
	}

	// Look for OFDM downstream
	ofdmDownstreamId := 0
	for idx, value := range iftypes {
		if value == 277 {
			ofdmDownstreamId = idx
		}
	}

	if ofdmDownstreamId > 0 {
		ofdms, err := getOfdmDownstreams(snmp, ofdmDownstreamId)
		if err != nil {
			return nil, err
		}
		data.OfdmDownstreams = ofdms
	}


	if data.DownStreams, data.DSPrimary, err = GetDownstreamChannels(snmp); err != nil {
		log.Printf(
			"Error getting downstream information from modem %s (%s): %s", macs, ip, err)
		data.Err = err
		return
	}

	if data.UpStreams, err = GetUpstreamChannels(snmp); err != nil {
		log.Printf(
			"Error getting upstream information from modem %s (%s): %s", macs, ip, err)
		data.Err = err
		return
	}

	data.QueryTime = int64(time.Now().Sub(start))
	return
}
