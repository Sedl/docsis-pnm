package types

import "github.com/sedl/docsis-pnm/internal/constants"

type CmStatus int

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
	if status < 1 || status > constants.CmStatusRegisteredBPIInitializing {
		return "Unknown"
	}

	return cmStatusString[status]
}
