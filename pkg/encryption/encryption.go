package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
)

func Decrypt(encrypted string, secretKey string, secretIv string) (string, error) {
	if len(encrypted) == 0 {
		return encrypted, nil
	}
	digest := getDigest(secretKey)
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

func getDigest(secretKey string) []byte {
	hash := md5.Sum([]byte(secretKey))
	return hash[:]
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
