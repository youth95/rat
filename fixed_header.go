package rat

import (
	"bytes"
	"encoding/binary"
)

const FixedHeaderSize = 5

type FixedHeader struct {
	MT byte
	ML int32
}

const (
	MTRequest = iota
	MTResponse
	MTHB
)

func (fixedHeader *FixedHeader) Bytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(fixedHeader.MT)
	err := binary.Write(buf, binary.LittleEndian, fixedHeader.ML)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func ParseFixedHeader(p []byte) *FixedHeader {
	var fixedHeader FixedHeader
	buf := bytes.NewBuffer(p)
	mt, err := buf.ReadByte()
	if err != nil {
		panic(err)
	}
	fixedHeader.MT = mt
	err = binary.Read(buf, binary.LittleEndian, &fixedHeader.ML)
	if err != nil {
		panic(err)
	}
	return &fixedHeader
}
