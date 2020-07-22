package modem

import (
	msnmp "github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/soniah/gosnmp"
)


func GetDownstreamChannels(snmp *gosnmp.GoSNMP) ([]types.DownstreamChannel, int32, error) {

	// docsIftypes.DownstreamChannelEntry
	res1, err := snmp.BulkWalkAll(".1.3.6.1.2.1.10.127.1.1.1.1")
	if err != nil {
		return nil, 0, err
	}

	// docsIfSignalQualityEntry
	res2, err := snmp.BulkWalkAll(".1.3.6.1.2.1.10.127.1.1.4.1")
	if err != nil {
		return nil, 0, err
	}

	results := append(res1, res2...)

	downstreams := make(map[int]*types.DownstreamChannel)

	var downstr *types.DownstreamChannel
	var ok bool
	var primary int32

	for _, result := range results {
		oid, idx := msnmp.SliceOID(result.Name)

		if downstr, ok = downstreams[idx]; !ok {
			downstr = &types.DownstreamChannel{Index: idx}
			downstreams[idx] = downstr
		}

		switch oid {

		// DOCS-IF-MIB::docsIfDownChannelId
		case ".1.3.6.1.2.1.10.127.1.1.1.1.1":
			downstr.ID, _ = msnmp.ToInt32(&result)

		// DOCS-IF-MIB::docsIfDownChannelFrequency
		case ".1.3.6.1.2.1.10.127.1.1.1.1.2":
			downstr.Freq, _ = msnmp.ToInt32(&result)
			if idx == 3 {
				primary = downstr.Freq
			}

		// DOCS-IF-MIB::docsIfDownChannelPower
		case ".1.3.6.1.2.1.10.127.1.1.1.1.6":
			if pow, err1 := msnmp.ToInt32(&result); err1 == nil {
				downstr.Power = pow
			}

		// DOCS-IF-MIB::docsIfSigQSignalNoise
		case ".1.3.6.1.2.1.10.127.1.1.4.1.5":
			if snr, err1 := msnmp.ToInt32(&result); err1 == nil {
				downstr.SNR = snr
			}

		// DOCS-IF-MIB::docsIfSigQMicroreflections
		case ".1.3.6.1.2.1.10.127.1.1.4.1.6":
			if refl, err1 := msnmp.ToInt32(&result); err1 == nil {
				downstr.Microrefl = refl * -1
			}

		// DOCS-IF-MIB::docsIfSigQExtUnerroreds
		case ".1.3.6.1.2.1.10.127.1.1.4.1.8":
			downstr.Unerroreds, _ = msnmp.ToUint64(&result)

		// DOCS-IF-MIB::docsIfSigQExtCorrecteds
		case ".1.3.6.1.2.1.10.127.1.1.4.1.9":
			downstr.Correcteds, _ = msnmp.ToUint64(&result)

		// DOCS-IF-MIB::docsIfSigQExtUncorrectables
		case ".1.3.6.1.2.1.10.127.1.1.4.1.10":
			downstr.Uncorrectables, _ = msnmp.ToUint64(&result)

		// DOCS-IF-MIB::docsIfDownChannelModulation
		case ".1.3.6.1.2.1.10.127.1.1.1.1.4":
			downstr.Modulation, _ = msnmp.ToInt32(&result)
		}
	}


	dslist := make([]types.DownstreamChannel, 0)
	for _, ds := range downstreams {
		dslist = append(dslist, *ds)
	}
	return dslist, primary, nil
}
