package rat

import (
	"bytes"
	"crypto/cipher"
)

func padding(src []byte, blockSize int) []byte {
	size := blockSize - len(src)%blockSize
	pad := bytes.Repeat([]byte{byte(size)}, size)
	return append(src, pad...)
}

func unPadding(src []byte) []byte {
	n := len(src)
	size := int(src[n-1])
	return src[:n-size]
}

func encryptAES(src []byte, block cipher.Block, key []byte, ) []byte {
	src = padding(src, block.BlockSize())
	mode := cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(src, src)
	return src
}

func decryptAES(src []byte, block cipher.Block, key []byte) []byte {
	mode := cipher.NewCBCDecrypter(block, key)
	mode.CryptBlocks(src, src)
	src = unPadding(src)
	return src
}
