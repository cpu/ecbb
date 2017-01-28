package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"fmt"
	"image"
)

// ecbEncrypt takes an input RGBA image and a fixed key and returns the image
// encrypted using AES 128 in ECB mode with a key derived from the key string
func ecbEncrypt(rgba image.RGBA, key string) (image.Image, error) {
	// Everything is an ECB Penguin if you squint hard enough
	penguin := image.NewRGBA(rgba.Bounds())

	// Turn the "key" string into a 16 byte AES key by computing the SHA1 sum and
	// slicing the first 16 bytes. This is a *terrible* key derivation strategy!
	// Don't do this unless you're writing a twitter bot that deliberately uses
	// bad crypto!
	hashFunc := sha1.New()
	hashFunc.Write([]byte(key))
	keyBytes := hashFunc.Sum(nil)[0:16]

	// Create an AES block cipher
	blockCipher, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}
	// Wrap it in the ECB block cipher mode
	e := newECBBlockCipher(blockCipher)

	srcBytes := pad(rgba.Pix, e.BlockSize())

	// Encrypt the image data into a buffer
	encryptedBytes := make([]byte, len(srcBytes))
	e.CryptBlocks(encryptedBytes, srcBytes)
	penguin.Pix = encryptedBytes
	return penguin, nil
}

// ecbBlockcipher is a struct wrapping a block cipher to operate in ECB mode
type ecbBlockcipher struct {
	cipher cipher.Block
}

// newECBBlockCipher wraps a `cipher.Block` instance to operate in ECB mode
// Note: It does *not* support decryption!
func newECBBlockCipher(cipher cipher.Block) cipher.BlockMode {
	return &ecbBlockcipher{
		cipher: cipher,
	}
}

// pad() will zero pad a plaintext message until it is a multiple of the block
// cipher blocksize. this is a terrible idea unless you're writing a shitty
// crypto twitter bot!
func pad(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(plaintext, padtext...)
}

// CryptBlocks is implemented to operate in ECB mode. It will panic if the input
// length isn't evenly divisble by the blocksize, or if the output buffer is
// smaller than the input buffer. For more information see
// https://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Electronic_Codebook_.28ECB.29
// Credit to https://gist.github.com/DeanThompson/17056cc40b4899e3e7f4 for the
// `CryptBlocks` implementation I based this on.
func (c *ecbBlockcipher) CryptBlocks(dst, src []byte) {
	bs := c.cipher.BlockSize()
	if len(src)%bs != 0 {
		panic(fmt.Sprintf("ecbb/ecbBlockcipher: input length (%d) not divisible by blocksize (%d)",
			len(src), bs))
	}
	if len(dst) < len(src) {
		panic(fmt.Sprintf("ecbb/ecbBlockcipher: output buffer length (%d) smaller than input length (%d)",
			len(dst), len(src)))
	}
	// While there is still input to read, loop
	for len(src) > 0 {
		// Encrypt one block from the src to the dest and advance the buffers
		c.cipher.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
}

// BlockSize is implemented to meet the `cipher.BlockMode` interface
func (c *ecbBlockcipher) BlockSize() int {
	return c.cipher.BlockSize()
}
