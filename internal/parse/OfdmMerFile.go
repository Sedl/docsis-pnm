package parse

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type OfdmRxMerFileHeader struct {
	Magic                      [4]byte `json:"-"`
	MajorVersion               uint8   `json:"-"`
	MinorVersion               uint8   `json:"-"`
	CaptureTime                uint32  `json:"capture_time"`
	ChannelId                  uint8   `json:"channel_id"`
	MacAddressRaw              [6]byte `json:"-"`
	SubcarrierZeroFrequency    uint32  `json:"subcarrier_zero_frequency"`
	FirstActiveSubcarrierIndex uint16  `json:"first_active_subcarrier_index"`
	SubcarrierSpacingKhz       uint8   `json:"subcarrier_spacing_khz"`
	DataLength                 uint32  `json:"- "`
}

type OfdmRxMerFileStruct struct {
	*OfdmRxMerFileHeader
	MacAddress string    `json:"mac_address"`
	MerDB      []float32 `json:"mer_db"`
}

var InvalidOfdmRxMerFile = errors.New("invalid OFDM Rx MER file")

var OfdmRxMerFileMagic = [4]byte{0x50, 0x4e, 0x4e, 0x40}

const MaxSubcarriers = 7800

func OfdmMerFile(data []byte) (*OfdmRxMerFileStruct, error) {
	reader := bytes.NewReader(data)

	header := &OfdmRxMerFileHeader{}
	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		return nil, err
	}

	if header.Magic == OfdmRxMerFileMagic {
		return nil, InvalidOfdmRxMerFile
	}

	if header.DataLength > MaxSubcarriers {
		return nil, errors.New(fmt.Sprintf("returned subcarrier count (%d) exceeds maximum subcarrier count (%d)", header.DataLength, MaxSubcarriers))
	}

	f := &OfdmRxMerFileStruct{
		OfdmRxMerFileHeader: header,
		MacAddress:          net.HardwareAddr(header.MacAddressRaw[:]).String(),
	}

	subcarriers := make([]byte, header.DataLength)
	readBytes, err := reader.Read(subcarriers)

	if readBytes < int(header.DataLength) {
		return nil, errors.New("unexpected end of data while reading subcarrier MER data")
	}

	merData := make([]float32, header.DataLength)
	for i, val := range subcarriers {
		merData[i] = float32(val) / 4
	}
	f.MerDB = merData
	return f, nil
}