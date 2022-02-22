package snmp

import (
	"fmt"
	"github.com/gosnmp/gosnmp"
	"log"
	"math"
	"strconv"
	"strings"
)

func SliceOID(oid string) (string, int) {
	last := strings.LastIndex(oid, ".")
	idx, err := strconv.Atoi(oid[last+1:])
	if err != nil {
		return "", -1
	}
	return oid[:last], idx
}

func ToString(result *gosnmp.SnmpPDU) (string, error) {
	switch result.Type {
	case gosnmp.OctetString:
		if result.Value == nil {
			return "", fmt.Errorf("snmp value is nil in snmp response (%s). expecting type integer (0x02)", result.Name)
		}

		return string(result.Value.([]byte)), nil
	default:
		return "", fmt.Errorf("invalid data type (%#x) in snmp response (%s). expecting type octetstring (0x04)", result.Type, result.Name)
	}

}
func ToInt64(result *gosnmp.SnmpPDU) (int64, error) {
	switch result.Type {
	case gosnmp.Integer:
		return int64(result.Value.(int)), nil
	case gosnmp.Counter32:
		return int64(result.Value.(uint)), nil
	case gosnmp.Counter64:
		res := result.Value.(uint64)
		if res > math.MaxInt64 {
			return 0, fmt.Errorf("integer too big to convert to int64 in snmp response (%s)", result.Name)
		}
		return int64(res), nil
	default:
		return 0, fmt.Errorf("invalid data type (%#x) in snmp response (%s). expecting type integer (0x02)",
			result.Type, result.Name)
	}
}

func ToUint(result *gosnmp.SnmpPDU, snmp *gosnmp.GoSNMP) (uint, error) {
	switch result.Type {
	case gosnmp.Gauge32:
		return result.Value.(uint), nil
	default:
		log.Printf("invalid data type (%#x) in snmp response from %s (%s). expecting type gauge32 (0x42).",
			result.Type, snmp.Target, result.Name)
		return 0, IntegerConversionError
	}
}

func ToUint32(result *gosnmp.SnmpPDU) (uint32, error) {
	switch result.Type {
	case gosnmp.TimeTicks:
		return result.Value.(uint32), nil
	case gosnmp.Gauge32, gosnmp.Counter32:
		return uint32(result.Value.(uint)), nil
	case gosnmp.Integer:
		return uint32(result.Value.(int)), nil
	default:
		t := int(result.Type)
		log.Printf("invalid data type (0x%x) in snmp response (%s). expecting type gauge32 (0x42).",
			t, result.Name)
		return 0, IntegerConversionError
	}
}

func ToUint32Map(results map[int]*gosnmp.SnmpPDU) (map[int]uint32, error) {
	converted := make(map[int]uint32)
	for idx, pdu := range results {
		value, err := ToUint32(pdu)
		if err != nil {
			return nil, err
		}
		converted[idx] = value
	}
	return converted, nil
}

func ToUint64(result *gosnmp.SnmpPDU) (uint64, error) {
	switch result.Type {
	case gosnmp.Counter64:
		return result.Value.(uint64), nil
	default:
		log.Printf("invalid data type (%#x) in snmp response (%s). expecting type Counter64 (0x46).",
			result.Type, result.Name)
		return 0, IntegerConversionError
	}
}

func ToInt(result *gosnmp.SnmpPDU) (int, error) {
	switch result.Type {
	case gosnmp.Integer:
		return result.Value.(int), nil
	default:
		log.Printf("invalid data type (%#x) in snmp response (%s). expecting type integer (0x42).",
			result.Type, result.Name)
		return 0, IntegerConversionError
	}
}

func ToInt32(result *gosnmp.SnmpPDU) (int32, error) {
	switch result.Type {
	case gosnmp.Gauge32:
		res := result.Value.(uint)
		if res > math.MaxInt32 {
			return 0, fmt.Errorf("integer too big to convert to int32 in snmp response (%s)", result.Name)
		}
		return int32(res), nil

	case gosnmp.Integer:
		res := result.Value.(int)
		// Maybe there is a better way?
		if res > math.MaxInt32 {
			return 0, fmt.Errorf("integer too big to convert to int32 in snmp response (%s)", result.Name)
		}
		if res < math.MinInt32 {
			return 0, fmt.Errorf("integer too small to convert to int32 in snmp response (%s)", result.Name)
		}
		return int32(res), nil
	default:
		return 0, fmt.Errorf("invalid data type (%#x) in snmp response (%s), expecting type integer (0x42)",
			result.Type, result.Name)
	}
}

