package rat

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)


func TestConnectTimeout(t *testing.T) {
	a := assert.New(t)
	_, err := ConnectTimeout("rat://0.0.0.0/hb", 1*time.Second)
	a.NotNil(err)
	_, err = ConnectTimeout(" r  ", 1*time.Second)
	a.NotNil(err)
	_, err = ConnectTimeout("mrat://0.0.0.0/hb", 1*time.Second)
	a.NotNil(err)
}
