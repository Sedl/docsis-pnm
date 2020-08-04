package types

// Upstream information from CMTS
type UpstreamModemCMTS struct {
	ModemId    ModemId
	UpstreamId int32
	PollTime   int32
	PowerRx    int32 // receive power in tenth dB
	SNR        int32 // signal to noise ratio in tenth dB
	Microrefl  int32
	Unerroreds int64
	Correcteds int64
	Erroreds   int64
}
