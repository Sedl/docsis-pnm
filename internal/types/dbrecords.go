package types

import (
	"fmt"
	"net"
	"time"
)

// CMTSRecord represents a CMTS database record
type CMTSRecord struct {
	Id                 uint32 `json:"cmts_id"`
	Hostname           string `json:"hostname"`
	SNMPCommunity      string `json:"snmp_community,omitempty"`
	SNMPModemCommunity string `json:"snmp_community_modem,omitempty"`
	Disabled           bool   `json:"disabled"`
	PollInterval       int32  `json:"poll_interval"`
}

type CMTSUpstreamRecord struct {
	ID          int32
	CMTSID      uint32
	SNMPIndex   int32
	Description string
	Freq        int32
	Alias       string
	AdminStatus int32
}

func (m *CMTSUpstreamRecord) String() string {
	return fmt.Sprintf("CMTSUpstreamRecord @%p (id=%d, cmts_id=%d, snmp_idx=%d, descr=%q, freq=%d, alias=%q)", m, m.ID, m.CMTSID, m.SNMPIndex, m.Description, m.Freq, m.Alias)
}

type CMTSUpstreamHistoryRecord struct {
	PollTime        time.Time
	UpstreamId      int32
	Unerroreds      int64
	Correcteds      int64
	Uncorrectrables int64
	Utilization     int32
	PktsBroadcast   int64
	PktsUnicast     int64
	Bytes           int64
	MER             float32
}

type ModemRecord struct {
	ID            uint64
	Mac           net.HardwareAddr
	SysDescr      string
	IP            net.IP
	CmtsId        uint32
	SnmpIndex     int32
	DocsisVersion uint32
	DSPrimary	  int32
	CmtsDsIndex	  int32
	CmtsUsIndex   int32
}
