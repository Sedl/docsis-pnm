package modem

import (
	snmp2 "github.com/sedl/docsis-pnm/internal/snmp"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/soniah/gosnmp"
	"strconv"
)

const (
	docsIf31CmDsOfdmChannelPowerCenterFrequency = ".1.3.6.1.4.1.4491.2.1.28.1.11.1.2"
	docsIf31CmDsOfdmChannelPowerRxPower = ".1.3.6.1.4.1.4491.2.1.28.1.11.1.3"

	// profile
	docsIf31CmDsOfdmProfileStatsConfigChangeCt = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.2"
	docsIf31CmDsOfdmProfileStatsTotalCodewords = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.3"
	docsIf31CmDsOfdmProfileStatsCorrectedCodewords = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.4"
	docsIf31CmDsOfdmProfileStatsUncorrectableCodewords = ".1.3.6.1.4.1.4491.2.1.28.1.10.1.5"
)


func getOfdmDownstreams(snmp *gosnmp.GoSNMP, downstreamIndex int)  (*types.OfdmDownstreamHistory, error) {

	oids := []string{
		docsIf31CmDsOfdmChannelPowerCenterFrequency,
		docsIf31CmDsOfdmChannelPowerRxPower,
	}

	appendix := "." + strconv.Itoa(downstreamIndex)
	subtree, err := snmp2.WalkSubtree(snmp, oids)
	if err != nil {
		return nil, err
	}

	oid := docsIf31CmDsOfdmChannelPowerCenterFrequency + appendix
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

	oid = docsIf31CmDsOfdmChannelPowerRxPower + appendix
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
	subtree, err = snmp2.WalkSubtree(snmp, oids)
	if err != nil {
		return nil, err
	}

	profiles := make(map[int]*types.OfdmDownstreamChannelProfileData)

	oid = docsIf31CmDsOfdmProfileStatsConfigChangeCt + appendix
	for idx, pdu := range subtree[oid] {
		chCount, err := snmp2.ToUint32(pdu)
		if err != nil {
			return nil, err
		}
		profiles[idx] = &types.OfdmDownstreamChannelProfileData{
			ChangeCount:      chCount,
		}
	}

	oid = docsIf31CmDsOfdmProfileStatsTotalCodewords + appendix
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

	oid = docsIf31CmDsOfdmProfileStatsCorrectedCodewords + appendix
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

	oid = docsIf31CmDsOfdmProfileStatsUncorrectableCodewords + appendix
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