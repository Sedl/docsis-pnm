package snmp

import "github.com/soniah/gosnmp"

func WalkSubtree(snmp *gosnmp.GoSNMP, oids []string) (map[string]map[int]*gosnmp.SnmpPDU, error) {
	subtree := make(map[string]map[int]*gosnmp.SnmpPDU)
	for _, oid := range oids {
		result, err := snmp.BulkWalkAll(oid)
		if err != nil {
			return nil, err
		}
		for _, pdu := range result {
			pdu1 := pdu
			oidr, idx := SliceOID(pdu1.Name)
			if pktMap, ok := subtree[oidr]; ok {
				pktMap[idx] = &pdu1
			} else {
				subtree[oidr] = make(map[int]*gosnmp.SnmpPDU)
				subtree[oidr][idx] = &pdu1
			}
		}
	}

	return subtree, nil
}

