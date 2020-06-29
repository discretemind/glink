package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/curve25519"
	"testing"
)

func TestSahred(t *testing.T) {

	var privateKey [32]byte
	rand.Read(privateKey[:])

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	fmt.Printf("\nAlice Private key (a):\t%x\n", privateKey)
	fmt.Printf("\nAlice Public key point (x co-ord):\t%x\n", publicKey)

	var privateKey2 [32]byte
	rand.Read(privateKey2[:])

	var publicKey2 [32]byte
	curve25519.ScalarBaseMult(&publicKey2, &privateKey2)

	fmt.Printf("\nBob Private key (b):\t%x\n", privateKey2)
	fmt.Printf("\nBob Public key point (x co-ord):\t%x\n", publicKey2)

	out1, _ := curve25519.X25519(privateKey[:], publicKey2[:])
	out2, _ := curve25519.X25519(privateKey2[:], publicKey[:])

	fmt.Printf("\nShared key (Alice):\t%x\n", out1)
	fmt.Printf("\nShared key (Bob):\t%x\n", out2)
}

func TestCards(t *testing.T) {
	key1 := GeneratePrivateKey()
	key2 := GeneratePrivateKey()

	original := "test secret message..."
	data, err := key1.Encrypt(key2.Public().Bytes(), []byte(original))
	assert.NoError(t, err)
	message, ok := key2.Decrypt(key1.Public().Bytes(), data)
	assert.True(t, ok)
	assert.Equal(t, original, string(message))
}

//
func TestSignature(t *testing.T) {
	key1 := GeneratePrivateKey()
	cert := key1.Certificate()

	original := []byte("test secret message...")
	sign := key1.Sign(original)

	ok := cert.Verify(original, sign)
	assert.True(t, ok)

	message2 := "test secret message"
	ok = cert.Verify([]byte(message2), sign)
	assert.False(t, ok)
}

func TestCArdIos(t *testing.T) {
	fmt.Println()
	data, _ := base64.StdEncoding.DecodeString("EHN8canU6Z9GWHSMSdnAjyW9G4KjfBSsu1qTWQi7jk0=")
	puData, _ := base64.StdEncoding.DecodeString("IHgHjsAUREOyWM2ivHX0xh0UgQRsY9ewSCeiKta4b0M=")
	fmt.Printf("%x\n", data)
	pub, err := getPublicKey(data[:])
	assert.NoError(t, err)
	fmt.Printf("%x\n", pub)
	fmt.Printf("%x\n", puData)

}
