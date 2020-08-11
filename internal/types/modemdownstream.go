package types

type DownstreamChannel struct {
	ID             int32  `json:"-"`
	Freq           int32  `json:"freq"`
	Power          int32  `json:"pwr"`
	SNR            int32  `json:"snr"`
	Microrefl      int32  `json:"mrefl"`
	Unerroreds     uint64 `json:"unerr"`
	Correcteds     uint64 `json:"corr"`
	Uncorrectables uint64 `json:"err"`
	Index          int    `json:"snmp_index,omitempty"`
	Modulation     int32  `json:"mod"`
}

type DownstreamChannelHistory struct {
	Timestamp   int64                `json:"ts"`
	Downstreams []*DownstreamChannel `json:"ds"`
}

type OfdmDownstreamHistory struct {
	ProfileData map[int]*OfdmDownstreamChannelProfileData
	Downstreams map[int]*OfdmDownstream
}

type OfdmDownstreamChannelProfileData struct {
	ChangeCount      uint32
	CwTotal          uint64
	CwCorrected      uint64
	CwUncorrectables uint64
}

type OfdmDownstream struct {
	Freq  uint32
	Power float32
}
