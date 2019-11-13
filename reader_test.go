package rat

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestEventEmitter_RemoveAllListener2(t *testing.T) {
	a := assert.New(t)
	buf := bytes.NewBufferString("abcdf")
	reader := bufio.NewReader(buf)
	var tmp []byte
	n,err := reader.Read(tmp)
	a.Nil(err)
	a.Equal(n,0)

	tmp = make([]byte,2)
	reader.Reset(buf)

	n,err = reader.Read(tmp)
	a.Nil(err)
	a.Equal(n,2)

	n,err = reader.Read(tmp)
	a.Nil(err)
	a.Equal(n,2)

	n,err = reader.Read(tmp)
	a.Nil(err)
	a.Equal(n,1)


	n,err = reader.Read(tmp)
	a.Equal(err,io.EOF)
	a.Equal(n,0)

}
