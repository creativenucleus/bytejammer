package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigning(t *testing.T) {
	cryptoPrv, err := NewCryptoPrivate()
	require.NoError(t, err, "Failed to create cryptoPrivate")
	//	require(cryptoPrv != nil, "Private key is nil")
	//	require(cryptoPrv.privateKey != nil, "Private key is nil")

	publicKeyPem, err := cryptoPrv.PublicKeyToPem()
	require.NoError(t, err, "Failed to get public key in PEM format")
	//	require(publicKey == nil, "Failed to create public key")

	cryptoPub, err := NewCryptoPublicFromPem(publicKeyPem)
	require.NoError(t, err, "Failed to create public key from PEM")
	//require(cryptoPub == nil, "Failed to create public key")

	challenge := "Some challeng to be signed then verified"
	challengeBytes := []byte(challenge)

	signed, err := cryptoPrv.Sign(challengeBytes)
	require.NoError(t, err, "Failed to sign")
	require.NotEmpty(t, signed, "Empty signed value")

	isVerified := cryptoPub.VerifySigned(challengeBytes, signed)
	require.Equal(t, true, isVerified, "Signature did not verify")
}
