package pouchkey

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/cloudflare/circl/sign/ed448"
	"testing"
)

func generateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func TestNew(t *testing.T) {
	seed, err := generateRandomBytes(ed448.SeedSize)
	if err != nil {
		t.Fatal(err)
	}

	key := New(seed)
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestNewHexKeys(t *testing.T) {
	seed, err := generateRandomBytes(ed448.SeedSize)
	if err != nil {
		t.Fatal(err)
	}

	hexPrivateKey, hexPublicKey := NewHexKeys(seed)

	_, err = ParseHexPrivateKey(hexPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseHexPrivateKey(hexPublicKey)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSignAndVerify(t *testing.T) {
	seed, err := generateRandomBytes(ed448.SeedSize)
	if err != nil {
		t.Fatal(err)
	}

	_, encodedPublicKey := NewHexKeys(seed)

	hexEncodedSeed := hex.EncodeToString(seed)

	challenge := "ThisIsTheChallenge"

	hexEncodedSignature, err := SignWithSeedAsHex(hexEncodedSeed, challenge)
	if err != nil {
		t.Fatal(err)
	}

	ok := VerifyWithPublicKey(encodedPublicKey, challenge, hexEncodedSignature)
	if !ok {
		t.Fatal("expected true")
	}
}
