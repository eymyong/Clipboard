package service

import (
	"bytes"

	"github.com/soyart/gfc/pkg/gfc"
)

type Password interface {
	EncryptBase64(password string) (string, error)
	DecryptBase64(ciphertext string) (string, error)
}

type PasswordImpl struct {
	keyAES string
}

func NewServicePassword(keyAES string) *PasswordImpl {
	return &PasswordImpl{
		keyAES: keyAES,
	}
}

func (s *PasswordImpl) EncryptBase64(password string) (string, error) {
	buf := bytes.NewBufferString(password)
	ciphertext, err := gfc.EncryptGCM(buf, []byte(s.keyAES))
	if err != nil {
		return "", err
	}

	ciphertextBase64, err := gfc.Encode(gfc.EncodingBase64, ciphertext)
	if err != nil {
		return "", err
	}

	return string(ciphertextBase64.Bytes()), nil
}

func (s *PasswordImpl) DecryptBase64(password string) (string, error) {
	buf := bytes.NewBufferString(password)
	ciphertext, err := gfc.Decode(gfc.EncodingBase64, buf)
	if err != nil {
		return "", err
	}

	plaintext, err := gfc.DecryptGCM(ciphertext, []byte(s.keyAES))
	if err != nil {
		return "", err
	}

	return string(plaintext.Bytes()), nil
}
