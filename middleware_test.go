package rat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAESMiddleware(t *testing.T) {
	a := assert.New(t)
	_, err := NewAESMiddleware([]byte("123"))
	a.NotNil(err)
}
