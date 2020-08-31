package modem

import (
	"errors"
	"fmt"
	"github.com/sedl/docsis-pnm/internal/constants"
	"github.com/soniah/gosnmp"
	"net"
)

var ErrorNoOfdmDownstream = errors.New("modem does not have an OFDM downstream channel")

// RequestPnmOfdmMerFile requests an OFDM MER file from a modem via TFTP. The tftp part is not handled here, it just
// sets the appropriate SNMP OIDs.
func (p *Poller) RequestPnmOfdmMerFile(filename string, tftpAddress net.IP) error {

	// TODO test with IPv6 address

	var addrType int
	if len(tftpAddress) > net.IPv4len {
		addrType = constants.AddrTypeIpv6
	} else {
		addrType = constants.AddrTypeIpv4
	}

	ofdmIdx, err := p.FindOfdmDownstreamIdx()
	if err != nil {
		return err
	}
	if ofdmIdx == 0 {
		return ErrorNoOfdmDownstream
	}

	ofdmIdxStr := fmt.Sprintf(".%d", ofdmIdx)
	pdus := []gosnmp.SnmpPDU{
		{
			Name:  constants.DocsPnmBulkDestIpAddrType + ".0",
			Type:  gosnmp.Integer,
			Value: addrType,
		},
		{
			Name:  constants.DocsPnmBulkDestIpAddr + ".0",
			Type:  gosnmp.OctetString,
			Value: []byte(tftpAddress),
		},
		{
			Name:  constants.DocsPnmCmDsOfdmRxMerFileName + ofdmIdxStr,
			Type:  gosnmp.OctetString,
			Value: filename,
		},
		{
			Name:  constants.DocsPnmCmDsOfdmRxMerFileEnable + ofdmIdxStr,
			Type:  gosnmp.Integer,
			Value: 1,
		},
	}
	_, err = p.snmp.Set(pdus)
	if err != nil {
		return err
	}

	return nil
}
