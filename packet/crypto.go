package packet

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

func Encrypt(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, data, nil)
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

func Decrypt(data []byte, priv *rsa.PrivateKey) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, data, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func GenerateKeys() (private *rsa.PrivateKey, public *rsa.PublicKey, err error) {
	private, _ = rsa.GenerateKey(rand.Reader, 4096)
	public = &private.PublicKey

	return private, public, nil
}
