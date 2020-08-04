package types

// Upstream information from CMTS
type UpstreamModemCMTS struct {
	ModemId    ModemId `json:"-"`
	UpstreamId int32   `json:"id"`
	PollTime   int32   `json:"-"`
	PowerRx    int32   `json:"pwr"` // receive power in tenth dB
	SNR        int32   `json:"snr"` // signal to noise ratio in tenth dB
	Microrefl  int32   `json:"mrefl"`
	Unerroreds int64   `json:"unerr"`
	Correcteds int64   `json:"corr"`
	Erroreds   int64   `json:"err"`
}
