package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"io"
)

func generateKeyWithCert() (publicKey [32]byte, cert [32]byte, privateKey [32]byte, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	fmt.Printf("%x\n", priv)
	fmt.Printf("%x\n", pub)

	copy(cert[:], pub)
	copy(privateKey[:], priv[:32])
	curve25519.ScalarBaseMult(&publicKey, &privateKey)
	//fmt.Printf("%x\n", publicKey)
	return
}

func getPublicKey(privateKey []byte) (publicKey [32]byte, err error) {
	var privCompressed [32]byte
	copy(privCompressed[:], privateKey[:])
	curve25519.ScalarBaseMult(&publicKey, &privCompressed)
	return
}

func sharedKey(secret, peer []byte) (shared []byte, err error) {
	return curve25519.X25519(secret[:], peer[:])
}

func sign(secret, message []byte) []byte {
	return ed25519.Sign(secret[:], message[:])
}

func verify(public, message, sign []byte) bool {
	return ed25519.Verify(public[:], message[:], sign[:])
}

func encryptShared(key []byte, message []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, message, nil), nil
}

func decryptShared(key []byte, ciphertext []byte) (result []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}
	return gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], nil)

}
