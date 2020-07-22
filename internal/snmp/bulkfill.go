package snmp

import (
	"github.com/soniah/gosnmp"
)

type RecordInterface interface {
	SNMPSetValue(pdu *gosnmp.SnmpPDU, create bool) error
}

// BulkFill takes a list of OIDs to Bulk walk over and create a list of SNMP records.
// A map of SNMP indices is createt from the first element in "oids", so make sure the first OID in the list has all the
// indices you want. New indices will be ignored.
func BulkFill(conn *gosnmp.GoSNMP,oids []string, records interface{ RecordInterface }) error {

	for i, oid := range oids {
		results, err := conn.BulkWalkAll(oid)
		if err != nil {
			return err
		}
		for _, result := range results {
			err = records.SNMPSetValue(&result, i == 0)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
