package pouchkey

import (
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"github.com/cloudflare/circl/sign/ed448"
)

func NewSeed() (string, error) {
	b, err := generateRandomByteArray(ed448.SeedSize)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NewChallenge(length int) (string, error) {
	b, err := generateRandomByteArray(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateRandomByteArray(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// New creates an ED448 Private Key using the given seed.
//
// The length of the seed should be ed448.SeedSize, else this will panic
func New(seed []byte) ed448.PrivateKey {
	return ed448.NewKeyFromSeed(seed)
}

// NewHexKeys creates an ED448 Private Key using the given seed.
//
// It encodes into hexadecimal format and returns both the keys.
func NewHexKeys(seed []byte) (string, string) {
	privateKey := New(seed)
	publicKey := privateKey.Public().(ed448.PublicKey)
	encodedPrivateKey := hex.EncodeToString(privateKey)
	encodedPublicKey := hex.EncodeToString(publicKey)
	return encodedPrivateKey, encodedPublicKey
}

func ParseHexPrivateKey(key string) (ed448.PrivateKey, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return keyBytes, nil
}

func ParseHexPublicKey(key string) (ed448.PublicKey, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return keyBytes, nil
}

func SignWithSeedAsHex(hexEncodedSeed string, challenge string) (string, error) {
	seed, err := hex.DecodeString(hexEncodedSeed)
	if err != nil {
		return "", err
	}
	privateKey := New(seed)
	signatureBytes, err := privateKey.Sign(rand.Reader, []byte(challenge), crypto.Hash(0))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signatureBytes), nil
}

func VerifyWithPublicKey(hexEncodedPublicKey string, challenge string, hexEncodedSignature string) bool {
	publicKey, err := ParseHexPublicKey(hexEncodedPublicKey)
	if err != nil {
		return false
	}
	signatureBytes, err := hex.DecodeString(hexEncodedSignature)
	if err != nil {
		return false
	}
	return ed448.VerifyAny(publicKey, []byte(challenge), signatureBytes, crypto.Hash(0))
}
