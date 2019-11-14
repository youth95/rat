package rat

import (
	"bytes"
	"crypto/cipher"
)

func padding(src []byte, blocksize int) []byte {
	padnum := blocksize - len(src)%blocksize
	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
	return append(src, pad...)
}

func unpadding(src []byte) []byte {
	n := len(src)
	unpadnum := int(src[n-1])
	return src[:n-unpadnum]
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
	src = unpadding(src)
	return src
}
