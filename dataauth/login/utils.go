package login

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
)

func DesencriptarPassword(textoCifradoBase64 string, ivBase64 string) (string, error) {
	claveBase64 := os.Getenv("DECODE_PASS_KEY")
	claveBytes, err := base64.StdEncoding.DecodeString(claveBase64)
	if err != nil {
		return "", fmt.Errorf("error al decodificar la clave Base64: %w", err)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		return "", fmt.Errorf("error al decodificar el IV Base64: %w", err)
	}

	textoCifradoBytes, err := base64.StdEncoding.DecodeString(textoCifradoBase64)
	if err != nil {
		return "", fmt.Errorf("error al decodificar el texto cifrado Base64: %w", err)
	}

	block, err := aes.NewCipher(claveBytes)
	if err != nil {
		return "", fmt.Errorf("error al crear el cifrador AES: %w", err)
	}

	if len(textoCifradoBytes) < aes.BlockSize {
		return "", fmt.Errorf("texto cifrado demasiado corto")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(textoCifradoBytes, textoCifradoBytes)

	// Despadding PKCS7
	textoPlanoBytes, err := pkcs7UnPadding(textoCifradoBytes)
	if err != nil {
		return "", fmt.Errorf("error al quitar el padding PKCS7: %w", err)
	}

	return string(textoPlanoBytes), nil
}

func pkcs7UnPadding(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("texto cifrado vacío")
	}
	paddingLen := int(ciphertext[len(ciphertext)-1])
	if paddingLen > len(ciphertext) || paddingLen == 0 {
		return nil, fmt.Errorf("padding PKCS7 inválido")
	}

	padding := ciphertext[len(ciphertext)-paddingLen:]
	for _, p := range padding {
		if p != byte(paddingLen) {
			return nil, fmt.Errorf("padding PKCS7 inconsistente")
		}
	}

	return ciphertext[:len(ciphertext)-paddingLen], nil
}
