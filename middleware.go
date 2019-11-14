package rat

import (
	"crypto/aes"
	"crypto/cipher"
)

type Middleware interface {
	Read(data []byte) ([]byte, error)
	Write(data []byte) ([]byte, error)
}

func (socket *Socket) Uses(ms []Middleware) {
	socket.middleware = append(socket.middleware, ms...)
}

func (socket *Socket) middlewareReadHandData(data []byte) ([]byte, error) {
	tmp := data
	for _, m := range socket.middleware {
		t, err := m.Read(tmp)
		if err != nil {
			return nil, err
		}
		tmp = t
	}
	return tmp, nil
}

func (socket *Socket) middlewareWriteHandData(data []byte) ([]byte, error) {
	tmp := data
	for _, m := range socket.middleware { // TODO need resort
		t, err := m.Write(tmp)
		if err != nil {
			return nil, err
		}
		tmp = t
	}
	return tmp, nil
}

type AESMiddleware struct {
	block cipher.Block
	key   []byte
}

func NewAESMiddleware(key []byte) (*AESMiddleware, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESMiddleware{block, key}, nil
}

func (aESMiddleware *AESMiddleware) Read(data [] byte) ([]byte, error) {

	return decryptAES(data, aESMiddleware.block, aESMiddleware.key), nil
}

func (aESMiddleware *AESMiddleware) Write(data [] byte) ([]byte, error) {
	return encryptAES(data, aESMiddleware.block, aESMiddleware.key), nil
}
