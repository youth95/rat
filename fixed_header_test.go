package rat

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type MockReaderError struct{}

func (mock *MockReaderError) Read(bs []byte) (int, error) {
	return 0, errors.New("err")
}

func TestWriteFixedHeaderFromMessage(t *testing.T) {
	a := assert.New(t)
	msg := make([]byte, 1024)
	wb := bytes.NewBuffer([]byte{})
	h, err := WriteFixedHeaderFromMessage(wb, 100*time.Microsecond, msg)
	a.Nil(err)
	a.NotEmpty(h)
	a.Equal(len(wb.Bytes()), FixedHeaderSize)

}

func TestReadFixedHeader(t *testing.T) {
	a := assert.New(t)
	msg := make([]byte, 1024)
	buf := bytes.NewBuffer([]byte{})
	h, err := WriteFixedHeaderFromMessage(buf, 100*time.Microsecond, msg)
	a.Nil(err)
	a.NotEmpty(h)
	ti, l, err := ReadFixedHeader(buf)
	a.Nil(err)
	a.Equal(l, int32(1024))
	a.NotEmpty(ti)

	var rErr MockReaderError
	ti, l, err = ReadFixedHeader(&rErr)
	a.NotNil(err)
	a.Empty(l)
	a.Nil(ti)
}
