package modem

import (
	msnmp "github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/soniah/gosnmp"
)



func GetUpstreamChannels(snmp *gosnmp.GoSNMP) ([]types.UpstreamChannel, error) {
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
		case ".1.3.6.1.2.1.10.127.1.1.2.1.1":
			// DOCS-IF-MIB::docsIfUpChannelId
			upstr.ID, _ = msnmp.ToInt32(&result)

		case ".1.3.6.1.2.1.10.127.1.1.2.1.2":
			// DOCS-IF-MIB::docsIfUpChannelFrequency
			upstr.Freq, _ = msnmp.ToInt32(&result)

		case ".1.3.6.1.2.1.10.127.1.1.2.1.3":
			// DOCS-IF-MIB::docsIfUpChannelWidth
			upstr.Width, _ = msnmp.ToInt32(&result)

		case ".1.3.6.1.2.1.10.127.1.1.2.1.4":
			// DOCS-IF-MIB::docsIfUpChannelModulationProfile
			upstr.ModulationProfile, _ = msnmp.ToUint32(&result)

		case ".1.3.6.1.2.1.10.127.1.1.2.1.6":
			// DOCS-IF-MIB::docsIfUpChannelTxTimingOffset
			upstr.TimingOffset, _ = msnmp.ToUint32(&result)
		}
	}

	uslist := make([]types.UpstreamChannel, 0)
	for _, ds := range upstreams {
		uslist = append(uslist, *ds)
	}
	return uslist, nil

}
