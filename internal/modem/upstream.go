package modem

import (
	msnmp "github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/soniah/gosnmp"
)


const (
	docsIf3CmStatusUsTxPower = ".1.3.6.1.4.1.4491.2.1.20.1.2.1.1"
	docsIfUpChannelId = ".1.3.6.1.2.1.10.127.1.1.2.1.1"
	docsIfUpChannelFrequency = ".1.3.6.1.2.1.10.127.1.1.2.1.2"
	docsIfUpChannelWidth = ".1.3.6.1.2.1.10.127.1.1.2.1.3"
	docsIfUpChannelModulationProfile = ".1.3.6.1.2.1.10.127.1.1.2.1.4"
	docsIfUpChannelTxTimingOffset = ".1.3.6.1.2.1.10.127.1.1.2.1.6"
	docsIfCmStatusTxPower = ".1.3.6.1.2.1.10.127.1.2.2.1.3"
)

func GetUpstreamChannels(snmp *gosnmp.GoSNMP, docsisVersion uint32) ([]types.UpstreamChannel, error) {
	results, err := snmp.BulkWalkAll(".1.3.6.1.2.1.10.127.1.1.2.1")
	if err != nil {
		return nil, err
	}

	upstreams := make(map[int]*types.UpstreamChannel)

	var upstr *types.UpstreamChannel
	var ok bool

	for _, result := range results {
		oid, idx := msnmp.SliceOID(result.Name)

		if upstr, ok = upstreams[idx]; !ok {
			upstr = &types.UpstreamChannel{Index: int32(idx)}
			upstreams[idx] = upstr
		}

		switch oid {
		case docsIfUpChannelId:
			upstr.ID, _ = msnmp.ToInt32(&result)

		case docsIfUpChannelFrequency:
			upstr.Freq, _ = msnmp.ToInt32(&result)

		case docsIfUpChannelWidth:
			upstr.Width, _ = msnmp.ToInt32(&result)

		case docsIfUpChannelTxTimingOffset:
			upstr.TimingOffset, _ = msnmp.ToUint32(&result)

		case docsIf3CmStatusUsTxPower:
			upstr.TxPower, _ = msnmp.ToInt32(&result)
		}
	}

	if docsisVersion >= DocsVer30 {
		// Get additional upstream metrics from DOCS-IF3 subtree
		results, err = snmp.BulkWalkAll(docsIf3CmStatusUsTxPower)

		for _, result := range results {

			_, idx := msnmp.SliceOID(result.Name)

			if upstr, ok = upstreams[idx]; !ok {
				continue
			}

			upstr.TxPower, _ = msnmp.ToInt32(&result)
		}
	} else if len(upstreams) > 0 {
	    // Before DOCSIS 3.0 we have only one upstream channel
	    results, err = snmp.BulkWalkAll(docsIfCmStatusTxPower)
		if err != nil {
			return nil, err
		}
		for _, v := range upstreams {
			v.TxPower, _ = msnmp.ToInt32(&results[0])
		}
		if len(results) > 0 {
			for _, v := range upstreams {
				v.TxPower, _ = msnmp.ToInt32(&results[0])
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
