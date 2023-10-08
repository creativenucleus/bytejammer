package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
)

type CryptoPrivate struct {
	privateKey *rsa.PrivateKey
}

// https://betterprogramming.pub/exploring-cryptography-in-go-signing-vs-encryption-f19534334ad
// Returns public key, private key, error
func newCryptoPrivate() (*CryptoPrivate, error) {
	c := CryptoPrivate{}

	var err error
	c.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func newCryptoPrivateFromString(keyRaw []byte) (*CryptoPrivate, error) {
	c := CryptoPrivate{}

	var err error
	c.privateKey, err = x509.ParsePKCS1PrivateKey(keyRaw)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c CryptoPrivate) privateKeyToRaw() []byte {
	data := x509.MarshalPKCS1PrivateKey(c.privateKey)

	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: data,
	}

	// Private key in PEM format
	return pem.EncodeToMemory(&privBlock)
}

func (c CryptoPrivate) publicKeyToRaw() ([]byte, error) {
	data, err := x509.MarshalPKIXPublicKey(&c.privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: data,
		},
	)

	return pem, nil
}

func (c *CryptoPrivate) MarshalJSON() ([]byte, error) {
	type Alias CryptoPrivate
	return json.Marshal(&struct {
		PrivateKey []byte `json:"privateKey"`
	}{
		PrivateKey: c.privateKeyToRaw(),
	})
}

func (c *CryptoPrivate) UnmarshalJSON(data []byte) error {
	type Alias CryptoPrivate
	aux := &struct {
		PrivateKey []byte `json:"privateKey"`
	}{}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(aux.PrivateKey)
	if block == nil {
		return errors.New("failed to parse PEM block containing the key")
	}

	c.privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	return nil
}

func hashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Make a hash of the data, then sign it
func (c CryptoPrivate) sign(data []byte) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hashData(data))
}

type CryptoPublic struct {
	publicKey *rsa.PublicKey
}

func newCryptoPublicFromString(keyString string) (*CryptoPublic, error) {
	c := CryptoPublic{}

	var err error
	c.publicKey, err = x509.ParsePKCS1PublicKey([]byte(keyString))
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c CryptoPublic) verifySigned(data []byte, signature []byte) (bool, error) {
	return rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hashData(data), signature) == nil, nil
}
