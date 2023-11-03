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
	"fmt"
)

type CryptoPrivate struct {
	privateKey *rsa.PrivateKey
}

// https://betterprogramming.pub/exploring-cryptography-in-go-signing-vs-encryption-f19534334ad
// Returns public key, private key, error
func newCryptoPrivate() (*CryptoPrivate, error) {
	c := CryptoPrivate{}

	var err error
	c.privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func newCryptoPrivateFromString(privKey string) (*CryptoPrivate, error) {
	c := CryptoPrivate{}

	var err error
	p, _ := pem.Decode([]byte(privKey))
	c.privateKey, err = x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c CryptoPrivate) privateKeyToPem() []byte {
	data := x509.MarshalPKCS1PrivateKey(c.privateKey)

	privBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: data,
	}

	// Private key in PEM format
	return pem.EncodeToMemory(&privBlock)
}

func (c CryptoPrivate) publicKeyToPem() ([]byte, error) {
	data, err := x509.MarshalPKIXPublicKey(c.privateKey.Public())
	if err != nil {
		return nil, err
	}

	pemData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: data,
		},
	)

	return pemData, nil
}

func (c *CryptoPrivate) MarshalJSON() ([]byte, error) {
	type Alias CryptoPrivate
	return json.Marshal(&struct {
		PrivateKey []byte `json:"privateKey"`
	}{
		PrivateKey: c.privateKeyToPem(),
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
	return rsa.SignPSS(rand.Reader, c.privateKey, crypto.SHA256, hashData(data), nil)
}

type CryptoPublic struct {
	publicKey *rsa.PublicKey
}

func newCryptoPublicFromPem(pemData []byte) (*CryptoPublic, error) {
	c := CryptoPublic{}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("bad key data: %s", "not PEM-encoded")
	}

	if block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("type was not public key: %s", block.Type)
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pubKey.(type) {
	case *rsa.PublicKey:
		c.publicKey = pubKey.(*rsa.PublicKey)
	default:
		return nil, fmt.Errorf("type was not known public key type")
	}

	return &c, nil
}

func (c CryptoPublic) verifySigned(data []byte, signature []byte) bool {
	err := rsa.VerifyPSS(c.publicKey, crypto.SHA256, hashData(data), signature, nil)
	return err == nil
}
