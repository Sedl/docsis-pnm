package types

type UpstreamChannel struct {
	ID                int32  `json:"up_id,omitempty"`
	Freq              int32  `json:"freq"`
	Width             int32  `json:"up_width_hz,omitempty"`
	TimingOffset      uint32 `json:"timing_offset"`
	Index             int32  `json:"snmp_index,omitempty"`
	TxPower           int32  `json:"tx_power"`
}
