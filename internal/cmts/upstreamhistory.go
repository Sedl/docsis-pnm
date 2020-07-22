package cmts

import (
	"fmt"
	"github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/soniah/gosnmp"
	"log"
	"strconv"
	"strings"
	"time"
)

type UpstreamHistoryRecordMap struct {
	PollTime time.Time
	Records map[int]*types.CMTSUpstreamHistoryRecord
}

func NewCMTSUpstreamHistoryRecordMap(pollTime time.Time) UpstreamHistoryRecordMap {
	return UpstreamHistoryRecordMap{
		PollTime:pollTime,
		Records: make(map[int]*types.CMTSUpstreamHistoryRecord),
	}
}

func getIndex(pdu *gosnmp.SnmpPDU) (string, int, error) {
	// some oddball of OID. The interface index is not at the end...🙄
	if strings.HasPrefix(pdu.Name, docsIfCmtsChannelUtUtilization) {
		oid := docsIfCmtsChannelUtUtilization
		idxs := strings.SplitN(pdu.Name[len(oid)+1:], ".", 2)
		if len(idxs) == 0 {
			return "", 0, fmt.Errorf("there is something broken with the \"docsIfCmtsChannelUtUtilization\" oid while trying to fill in CMTSUpstreamHistoryRecord")
		}
		idx, err := strconv.Atoi(idxs[0])
		if err != nil {
			return "", 0, err
		}
		return oid, idx, nil
	} else {
		oid, idx := snmp.SliceOID(pdu.Name)
		return oid, idx, nil
	}
}
func (hist UpstreamHistoryRecordMap) SNMPSetValue(pdu *gosnmp.SnmpPDU, create bool) error {
	oid, idx, err := getIndex(pdu)
	if err != nil {
		log.Println(err)
		return err
	}

	val, found := hist.Records[idx]
	if ! found {
		if ! create {
			return nil
		}
		val = &types.CMTSUpstreamHistoryRecord{PollTime: hist.PollTime}
		hist.Records[idx] = val
	}

	switch oid {
	case docsIfSigQExtCorrecteds:
		val.Correcteds, err = snmp.ToInt64(pdu)
	case docsIfSigQExtUnerroreds:
		val.Unerroreds, err = snmp.ToInt64(pdu)
	case docsIfSigQExtUncorrectables:
		val.Uncorrectrables, err = snmp.ToInt64(pdu)
	case docsIfCmtsChannelUtUtilization:
		val.Utilization, err = snmp.ToInt32(pdu)
	case ifHCInBroadcastPkts:
		val.PktsBroadcast, err = snmp.ToInt64(pdu)
	case ifHCInUcastPkts:
		val.PktsUnicast, err = snmp.ToInt64(pdu)
	case ifHCInOctets:
		val.Bytes, err = snmp.ToInt64(pdu)
	default:
		return fmt.Errorf("unknown oid %q while trying to fill in CMTSUpstreamHistoryRecord", oid)
	}
	if err != nil {
		return err
	}
	return nil
}

var upstreamHistoryOIDs = []string {
	docsIfSigQExtUnerroreds,
	docsIfSigQExtCorrecteds,
	docsIfSigQExtUncorrectables,
	docsIfCmtsChannelUtUtilization,
	ifHCInBroadcastPkts,
	ifHCInUcastPkts,
	ifHCInOctets,
}

func (cmts *Cmts) pollUpstreamHistory() (map[int]*types.CMTSUpstreamHistoryRecord, error) {

	start := time.Now()
	records := NewCMTSUpstreamHistoryRecordMap(start)

	log.Printf("debug: starting to collect upstream history data from CMTS %q", cmts.dbRec.Hostname)

	err := snmp.BulkFill(cmts.Snmp, upstreamHistoryOIDs, records)
	if err != nil {
		return nil, err
	}

	log.Printf("debug: collecting upstream history data from CMTS %q done\n", cmts.dbRec.Hostname)


	return records.Records, nil
}
