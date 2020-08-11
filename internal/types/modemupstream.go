package types

type UpstreamChannel struct {
	ID                int32  `json:"upstream_id"`
	Freq              int32  `json:"freq"`
	Width             int32  `json:"channel_width,omitempty"`
	TimingOffset      uint32 `json:"timing_offset"`
	Index             int32  `json:"snmp_index"`
	TxPower           int32  `json:"tx_power"`
}
