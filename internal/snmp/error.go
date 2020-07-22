package snmp

import (
	"errors"
)

var (
	IntegerConversionError = errors.New("invalid data type. can't convert to Int")
	StringConversionError = errors.New("invalid data type. can't convert to String")
	NilError = errors.New("SNMP PDU returned nil value")
)
