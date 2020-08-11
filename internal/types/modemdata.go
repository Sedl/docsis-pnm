package types

import (
	"net"
)

type ModemData struct {
	Sysdescr        string                 `json:"sysdescr"`
	Hostname        string                 `json:"hostname"`
	QueryTime       int64                  `json:"query_time_ns"`
	Timestamp       int64                  `json:"timestamp"`
	DownStreams     []DownstreamChannel    `json:"ds,omitempty"`
	UpStreams       []UpstreamChannel      `json:"us,omitempty"`
	Mac             net.HardwareAddr       `json:"-"`
	CmtsDbId        int32                  `json:"-"`
	DbModemId       uint64                 `json:"-"`
	Err             error                  `json:"err,omitempty"`
	SnmpIndex       int32                  `json:"-"`
	DocsisVersion   uint32                 `json:"docsis_version"`
	DSPrimary       int32                  `json:"ds_primary"`
	OfdmDownstreams *OfdmDownstreamHistory `json:"-"`
	Uptime          uint32                 `json:"uptime"`
}

type ModemPollRequest struct {
	Hostname string // hostname or IP
	// CmtsId The database ID where the modem is currently active. This is here so the database gets an update of the
	// modems location. The location can change because customers tend to carry modems around...
	CmtsId    int32
	Mac       net.HardwareAddr
	SnmpIndex int32
	Community string
}
