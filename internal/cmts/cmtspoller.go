package cmts

import (
	"fmt"
	"github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/gosnmp/gosnmp"
	"log"
	"time"
)

// DOCS-IF-MIB::
const docsIfUpChannelFrequency = ".1.3.6.1.2.1.10.127.1.1.2.1.2"
const docsIfUpChannelModulationProfile = ".1.3.6.1.2.1.10.127.1.1.2.1.4"
const docsIfSigQExtUnerroreds = ".1.3.6.1.2.1.10.127.1.1.4.1.8"
const docsIfSigQExtCorrecteds = ".1.3.6.1.2.1.10.127.1.1.4.1.9"
const docsIfSigQExtUncorrectables = ".1.3.6.1.2.1.10.127.1.1.4.1.10"
const docsIfCmtsChannelUtUtilization = ".1.3.6.1.2.1.10.127.1.3.9.1.3"

// IF-MIB::
const ifAlias = ".1.3.6.1.2.1.31.1.1.1.18"
const ifDescr = ".1.3.6.1.2.1.2.2.1.2"
const ifAdminStatus = ".1.3.6.1.2.1.2.2.1.7"
const ifHCInBroadcastPkts = ".1.3.6.1.2.1.31.1.1.1.9"
const ifHCInUcastPkts = ".1.3.6.1.2.1.31.1.1.1.7"
const ifHCInOctets = ".1.3.6.1.2.1.31.1.1.1.6"


var upstreamOIDs = []string{
	ifAlias,
	ifDescr,
	ifAdminStatus,
}

var upstreamConverters = map[string]func (*gosnmp.SnmpPDU, *types.CMTSUpstreamRecord) error {

	ifAlias: func(pdu *gosnmp.SnmpPDU, record *types.CMTSUpstreamRecord) (err error) {
		record.Alias, err = snmp.ToString(pdu)
		return
	},

	ifDescr: func(pdu *gosnmp.SnmpPDU, record *types.CMTSUpstreamRecord) (err error) {
		record.Description, err = snmp.ToString(pdu)
		return
	},

	ifAdminStatus: func(pdu *gosnmp.SnmpPDU, record *types.CMTSUpstreamRecord) (err error) {
		record.AdminStatus, err = snmp.ToInt32(pdu)
		return
	},
}

// pollUpstreams collects upstream information from the CMTS and updates the database accordingly
func (cmts *Cmts) pollUpstreams() (map[int]*types.CMTSUpstreamRecord, error) {
	start := time.Now()
	log.Printf("Collecting upstream information from CMTS %s\n", cmts.dbRec.Hostname)

	upstreams := make(map[int]*types.CMTSUpstreamRecord)

	// create initial list of upstreams using docsIfUpChannelFrequency
	results, err := cmts.Snmp.BulkWalkAll(docsIfUpChannelFrequency)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		oid, idx := snmp.SliceOID(result.Name)
		if oid != docsIfUpChannelFrequency { continue }
		freq, err := snmp.ToInt32(&result)
		if err != nil {
			log.Printf("Error while collecting upstream information data from %q: %s", cmts.dbRec.Hostname, err)
			continue
		}
		upstreams[idx] = &types.CMTSUpstreamRecord{SNMPIndex: int32(idx), Freq:freq, CMTSID:cmts.dbRec.Id}
	}

	// collect other upstream data
	for _, oid := range upstreamOIDs {
		converter, ok := upstreamConverters[oid]
		if ! ok {
			log.Printf("Error: Can't find data converter for OID %q", oid)
			continue
		}

		// collect
		results, err := cmts.Snmp.BulkWalkAll(oid)
		if err != nil {
			return nil, err
		}

		// fill in
		for _, result := range results {
			oidPolled, idx := snmp.SliceOID(result.Name)
			if oidPolled != oid {continue}

			val, found := upstreams[idx]
			if ! found {continue}

			err = converter(&result, val)
			// Don't make a conversion error a fatal one
			if err != nil {
				log.Printf("Error on device %q: %s", cmts.dbRec.Hostname, err)
			}
		}
	}

	log.Printf("Finished collecting upstream information from CMTS %s, took %s", cmts.dbRec.Hostname, time.Since(start))

	return upstreams, nil
}

func (cmts *Cmts) snmpCollector() error {

	// collect upstream information first to populate upstream cache
	// This cache is used for the SNMP index to database id lookups we will need for inserting the
	// upstream performance data (history)
	log.Printf("debug: polling upstreams for CMTS %s", cmts.dbRec.Hostname)
	upstreams, err := cmts.pollUpstreams()
	if err != nil {
		return fmt.Errorf("error on device %q while collecting upstream information: %s", cmts.dbRec.Hostname, err)
	}

	log.Printf("debug: updating upstreams for CMTS %s", cmts.dbRec.Hostname)
	err = cmts.DBBackend.UpdateCmtsUpstreams(upstreams)
	if err != nil {
		return fmt.Errorf("error on device %q while updating database for upstream information: %s", cmts.dbRec.Hostname, err)
	}

	// update upstream cache
	for _, upstr := range upstreams {
		 cmts.upstreamCache.Add(upstr)
	}

	log.Printf("debug: polling upstream history for CMTS %s", cmts.dbRec.Hostname)
	upshist, err := cmts.pollUpstreamHistory()
	if err != nil {
		return fmt.Errorf("error on device %q while collecting upstream history information: %s", cmts.dbRec.Hostname, err)
	}

	for idx, hist := range upshist {
		cached := cmts.upstreamCache.GetByIndex(idx)
		if cached == nil {
			continue
		}
		if cached.AdminStatus != 1 {
			// interface is not up
			continue
		}
		hist.UpstreamId = cached.ID
		err = cmts.DBBackend.InsertCMTSUpstreamHistory(hist)
		if err != nil {
			log.Println(err)
		}
	}

	log.Printf("debug: fetching modems from CMTS %s", cmts.dbRec.Hostname)
	err = cmts.updateModemList()
	if err != nil {
		log.Printf("Can't fetch modem list for CMTS %s. Reason: %s\n", cmts.dbRec.Hostname, err)
	}

	log.Printf("debug: fetching detailed modem upstream information")
	err, mdmUpstreams := cmts.pollModemUpstreams()
	if err != nil {
		return err
	}

	i := 0
	for _, mdm := range mdmUpstreams {
		for _, us := range mdm {
			// prevents inserting of erroneous records and offline modems (ModemId == 0)
		    if us.UpstreamId == 0 || us.ModemId == 0 {
		    	continue
			}
			cmts.dbSyncer.InsertCmtsModemUpstream(us)
			i++
		}
	}

	log.Printf("debug: inserted %d CMTS upstream records", i)

	return nil
}

func (cmts *Cmts) GoCMTSPoller() {
	// var start time.Time

	pollInterval := time.Duration(cmts.ValueOfCmtsPollInterval()) * time.Second

	log.Printf("Starting CMTS poller for %q with a poll interval of %s", cmts.dbRec.Hostname, pollInterval)


	err := cmts.snmpCollector()
	if err != nil {
		log.Printf("error: data collection failed for CMTS %s:%s\n", cmts.ValueOfHostname(), err.Error())
	}

	ticker := time.NewTicker(pollInterval)
	for {
		select {
		case _, ok := <-cmts.stopChannel:
			if ! ok {
				log.Printf("Stopping CMTS poller for %q", cmts.dbRec.Hostname)
				return
			}
		case <-ticker.C:
			err := cmts.snmpCollector()
			if err != nil {
				log.Printf("error: data collection failed for CMTS %s:%s\n", cmts.ValueOfHostname(), err.Error())
			}
		}
	}
}
