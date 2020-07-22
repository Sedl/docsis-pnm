package cmts

import (
	"net"
)

func (cmts *Cmts) GetModemCommunity(_ net.HardwareAddr) string {
	comm := cmts.dbRec.SNMPModemCommunity
	if comm == "" {
		comm = cmts.config.Snmp.Community
	}
	return comm
}
