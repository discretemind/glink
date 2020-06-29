package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/curve25519"
)

type PrivateKey [64]byte
type PublicKey [32]byte
type Certificate [32]byte

func CertificateFromString(value string) (res Certificate) {
	data, _ := base64.StdEncoding.DecodeString(value)
	copy(res[:], data)
	return
}

func GeneratePrivateKey() (res PrivateKey) {
	pubData, privData, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	copy(res[:32], privData[:])
	copy(res[32:], pubData[:])
	return
}

func PrivateKeyFromBase64(value string) (res PrivateKey, err error) {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return res, err
	}

	if len(data) != 64 {
		return res, errors.New("wrong key size. Should be 64")
	}

	copy(res[:64], data[:64])
	return
}

func (pk PrivateKey) Base64() string {
	return base64.StdEncoding.EncodeToString(pk[:])
}

func (pk PrivateKey) String() string {
	return pk.Base64()
}

func (pk PrivateKey) Certificate() (res Certificate) {
	copy(res[:], pk[32:])
	return
}

func (pk PrivateKey) Public() (res PublicKey) {
	pub := [32]byte{}
	pkData := [32]byte{}
	copy(pkData[:], pk[:32])
	curve25519.ScalarBaseMult(&pub, &pkData)
	copy(res[:32], pub[:32])
	return
}

func (pk PrivateKey) Encrypt(peerKey [32]byte, message []byte) (result []byte, err error) {
	lim := 300
	buf := message[:]
	shared, err := sharedKey(pk[:32], peerKey[:])
	if err != nil {
		return nil, err
	}
	var chunk []byte
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		encrypted, _ := encryptShared(shared, chunk)
		result = append(result, encrypted...)
	}
	if len(buf) > 0 {
		encrypted, _ := encryptShared(shared, buf)
		result = append(result, encrypted...)
	}
	return
}

func (pk PrivateKey) Decrypt(peerKey [32]byte, message []byte) (result []byte, ok bool) {
	lim := 340
	buf := message[:]
	shared, err := sharedKey(pk[:32], peerKey[:])
	if err != nil {
		return nil, false
	}
	var chunk []byte
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		decrypted, err := decryptShared(shared, chunk)
		if err != nil {
			return nil, false
		}
		result = append(result, decrypted...)
	}
	if len(buf) > 0 {
		decrypted, err := decryptShared(shared, buf)
		if err != nil {
			return nil, false
		}
		result = append(result, decrypted...)
	}
	return result, true
}

func (pk PrivateKey) Sign(message []byte) (signature []byte) {
	return sign(pk[:], message)
}

func (pk PrivateKey) Bytes() []byte {
	return pk[:]
}

func (crt Certificate) Verify(message []byte, signature []byte) (res bool) {
	return verify(crt[:], message, signature)
}

func (crt Certificate) String() string {
	return base64.StdEncoding.EncodeToString(crt[:])
}

func (pub PublicKey) Bytes() []byte {
	return pub[:]
}
func (pub PublicKey) String() string {

	return base64.StdEncoding.EncodeToString(pub[:])
}

func (pub PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", base64.URLEncoding.EncodeToString(pub[:]))), nil
}

func (pub *PublicKey) UnmarshalJSON(b []byte) error {
	str := string(b)
	data, err := base64.URLEncoding.DecodeString(str[1 : len(str)-1])
	if err != nil {
		return err
	}
	copy((*pub)[:], data[:])
	return err
}

func (pk PrivateKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", base64.URLEncoding.EncodeToString(pk[:]))), nil
}

func (pk *PrivateKey) UnmarshalJSON(b []byte) error {
	str := string(b)
	data, err := base64.URLEncoding.DecodeString(str[1 : len(str)-1])
	if err != nil {
		return err
	}
	copy((*pk)[:], data[:])
	return err
}

func (crt Certificate) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", base64.URLEncoding.EncodeToString(crt[:]))), nil
}

func (crt *Certificate) UnmarshalJSON(b []byte) error {
	str := string(b)
	data, err := base64.URLEncoding.DecodeString(str[1 : len(str)-1])
	if err != nil {
		return err
	}
	copy((*crt)[:], data[:])
	return err
}
