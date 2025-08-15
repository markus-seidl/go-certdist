package client

import (
	"go-certdist/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAgeKeys(t *testing.T) {
	privateKey, publicKey := common.NewAgeTestKey(t)

	t.Run("valid keys", func(t *testing.T) {
		config := &common.ClientModeConfig{
			AgeKey: common.AgeKeyConfig{
				PrivateKey: privateKey,
				PublicKey:  publicKey,
			},
		}
		assert.NoError(t, validateAgeKeys(config))
	})

	t.Run("public key auto-creation", func(t *testing.T) {
		config := &common.ClientModeConfig{
			AgeKey: common.AgeKeyConfig{
				PrivateKey: privateKey,
				PublicKey:  "", // Empty public key
			},
		}
		assert.NoError(t, validateAgeKeys(config))
		assert.Equal(t, publicKey, config.AgeKey.PublicKey)
	})
}

func TestValidateConnectionDetails(t *testing.T) {
	t.Run("add https prefix", func(t *testing.T) {
		config := &common.ClientModeConfig{
			ConnectionDetails: common.ClientConnectionConfig{
				Server: "example.com",
			},
		}
		assert.NoError(t, validateConnectionDetails(config))
		assert.Equal(t, "https://example.com", config.ConnectionDetails.Server)
	})

	t.Run("remove trailing slash", func(t *testing.T) {
		config := &common.ClientModeConfig{
			ConnectionDetails: common.ClientConnectionConfig{
				Server: "https://example.com/",
			},
		}
		assert.NoError(t, validateConnectionDetails(config))
		assert.Equal(t, "https://example.com", config.ConnectionDetails.Server)
	})

	t.Run("http is allowed", func(t *testing.T) {
		config := &common.ClientModeConfig{
			ConnectionDetails: common.ClientConnectionConfig{
				Server: "http://example.com",
			},
		}
		assert.NoError(t, validateConnectionDetails(config))
		assert.Equal(t, "http://example.com", config.ConnectionDetails.Server)
	})
}

func TestValidateCertificates(t *testing.T) {
	t.Run("valid certificate config", func(t *testing.T) {
		config := &common.ClientModeConfig{
			Certificate: []common.CertificateConfig{
				{
					Domain:    "example.com",
					Directory: "/tmp/certs",
				},
			},
		}
		assert.NoError(t, validateCertificates(config))
	})
}
