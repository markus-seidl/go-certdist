package common

import (
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/require"
)

// NewAgeTestKey generates a new age key pair for testing purposes.
func NewAgeTestKey(t *testing.T) (privateKey string, publicKey string) {
	t.Helper()
	identity, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	return identity.String(), identity.Recipient().String()
}
