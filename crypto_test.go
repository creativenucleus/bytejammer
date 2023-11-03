package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigning(t *testing.T) {
	cryptoPrv, err := newCryptoPrivate()
	require.NoError(t, err, "Failed to create cryptoPrivate")
	//	require(cryptoPrv != nil, "Private key is nil")
	//	require(cryptoPrv.privateKey != nil, "Private key is nil")

	publicKeyPem, err := cryptoPrv.publicKeyToPem()
	require.NoError(t, err, "Failed to get public key in PEM format")
	//	require(publicKey == nil, "Failed to create public key")

	cryptoPub, err := newCryptoPublicFromPem(publicKeyPem)
	require.NoError(t, err, "Failed to create public key from PEM")
	//require(cryptoPub == nil, "Failed to create public key")

	challenge := "Some challeng to be signed then verified"
	challengeBytes := []byte(challenge)

	signed, err := cryptoPrv.sign(challengeBytes)
	require.NoError(t, err, "Failed to sign")
	require.NotEmpty(t, signed, "Empty signed value")

	isVerified := cryptoPub.verifySigned(challengeBytes, signed)
	require.Equal(t, true, isVerified, "Signature did not verify")
}
