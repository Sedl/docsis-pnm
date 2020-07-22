package types

import (
	"net"
)

type ModemData struct {
	Sysdescr        string
	Hostname        string
	QueryTime       int64
	Timestamp       int64
	DownStreams     []DownstreamChannel
	UpStreams       []UpstreamChannel
	Mac             net.HardwareAddr
	CmtsDbId        uint32
	DbModemId       uint64
	Err             error
	Errors          []string `json:"errors"`
	SnmpIndex       int32
	DocsisVersion   uint32
	DSPrimary       int32
	OfdmDownstreams *OfdmDownstreamHistory
}

type ModemPollRequest struct {
	Hostname string // hostname or IP
	// CmtsId The database ID where the modem is currently active. This is here so the database gets an update of the
	// modems location. The location can change because customers tend to carry modems around...
	CmtsId    uint32
	Mac       net.HardwareAddr
	SnmpIndex int32
	Community string
}
