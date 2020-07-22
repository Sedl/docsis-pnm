package types

type UpstreamChannel struct {
	ID                int32 `json:"up_id,omitempty"`
	Freq              int32 `json:"up_freq_hz"`
	Width             int32 `json:"up_width_hz,omitempty"`
	ModulationProfile uint32  `json:"mod_profile"`
	TimingOffset      uint32  `json:"timing_offset"`
	Index             int32   `json:"snmp_index,omitempty"`
}

type UpstreamChannelHistory struct {
	Timestamp int64
	Upstreams []*UpstreamChannel
}
