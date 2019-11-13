package rat

import (
	"encoding/binary"
	"io"
	"time"
)

const FixedHeaderSize = 12

func WriteFixedHeaderFromMessage(w io.Writer, timeout time.Duration, msg []byte) (int64, error) {
	time.Sleep(1 * time.Microsecond) // 确保time.Now()是唯一的
	tl := time.Now().Add(timeout).Unix()
	err := binary.Write(w, binary.LittleEndian, tl)
	if err != nil {
		return 0, err
	}
	err = binary.Write(w, binary.LittleEndian, int32(len(msg)))
	if err != nil {
		return 0, err
	}
	return tl, nil
}

func ReadFixedHeader(r io.Reader) (*time.Time, int32, error) {
	var t time.Time
	var ut int64
	var l int32
	err := binary.Read(r, binary.LittleEndian, &ut)

	if err != nil {
		return nil, 0, err
	}
	err = binary.Read(r, binary.LittleEndian, &l)
	if err != nil {
		return nil, 0, err
	}
	t = time.Unix(ut, 0)
	return &t, l, nil
}
