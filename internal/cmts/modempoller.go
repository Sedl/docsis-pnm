package cmts

func (cmts *Cmts) GetModemCommunity() string {
	comm := cmts.dbRec.SNMPModemCommunity
	if comm == "" {
		comm = cmts.config.Snmp.Community
	}
	return comm
}
