package types

import (
	"net"
)

type ModemInfo struct {
	Index          int32            `json:"snmp_index"`
	MAC            net.HardwareAddr `json:"mac_address,string"`
	IP             net.IP           `json:"ip_address"`
	DownIfIndex    int32              `json:"downstr_snmp_index"`
	UpIfIndex      int32              `json:"upstr_snmp_index"`
	PowerRx        int              `json:"power_dbmv"`
	TimingOffset   uint             `json:"timing_offset"`
	Status         CmStatus         `json:"status"`
	Unerroreds     uint64           `json:"cw_unerroreds"`
	Correcteds     uint64           `json:"cw_correcteds"`
	Uncorrectables uint64           `json:"cw_uncorrectables"`
	CmtsDbId       uint32           `json:"-"`
	Timestamp      int64            `json:"-"`
}

type UpstreamCMTS struct {
	Timestamp int64
	PowerRx int
	Status int
	Unerroreds uint64
	Correcteds uint64
	Uncorrectables uint64
}
