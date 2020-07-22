package cmts

import (
	"net"
)

func (cmts *Cmts) GetModemCommunity(mac net.HardwareAddr) string {
	comm := cmts.dbRec.SNMPModemCommunity
	if comm == "" {
		comm = "public"
	}
	return comm
}

