package cmts

type CmStatus int

const (
	CmStatusOther                     = 1
	CmStatusRanging                   = 2
	CmStatusRangingAborted            = 3
	CmStatusRangingComplete           = 4
	CmStatusIpComplete                = 5
	CmStatusRegistrationComplete      = 6
	CmStatusAccessDenied              = 7
	CmStatusOperational               = 8
	CmStatusRegisteredBPIInitializing = 9
)

var cmStatusString = []string{
	"Unknown",
	"Other",
	"Ranging",
	"RangingAborted",
	"RangingComplete",
	"IpComplete",
	"RegistrationComplete",
	"AccessDenied",
	"Operational",
	"RegisteredBPIInitializing",
}

func (status CmStatus) String() string {
	if status < 1 || status > CmStatusRegisteredBPIInitializing {
		return "Unknown"
	}

	return cmStatusString[status]
}
