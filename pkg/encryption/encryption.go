package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
)

const (
	secretKey = "f1993e87-8db6-4748-9d86-a1e652f6"
	secretIv  = "abcdef1234567890"
)

func Encrypt(plainText string) (string, error) {
	if len(plainText) == 0 {
		return plainText, nil
	}
	digest := getDigest()
	block, err := aes.NewCipher(digest)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	input := pkcs5Padding([]byte(plainText), block.BlockSize())
	ivBytes := []byte(secretIv)[:block.BlockSize()]
	cipherText := make([]byte, len(input))
	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(cipherText, input)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(encrypted string) (string, error) {
	if len(encrypted) == 0 {
		return encrypted, nil
	}
	digest := getDigest()
	block, err := aes.NewCipher(digest)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	cipherText, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	if len(cipherText)%block.BlockSize() != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	ivBytes := []byte(secretIv)[:block.BlockSize()]
	decrypted := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(decrypted, cipherText)
	decrypted, err = pkcs5Unpadding(decrypted, block.BlockSize())
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func getDigest() []byte {
	hash := md5.Sum([]byte(secretKey))
	return hash[:]
}

func pkcs5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func pkcs5Unpadding(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("invalid padding")
	}
	padding := int(data[length-1])
	if padding > blockSize || padding == 0 {
		return nil, errors.New("invalid padding size")
	}
	for i := 0; i < padding; i++ {
		if data[length-1-i] != byte(padding) {
			return nil, errors.New("invalid padding content")
		}
	}
	return data[:length-padding], nil
}
