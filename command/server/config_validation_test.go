package server

import (
	"go-certdist/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateServerDetails(t *testing.T) {
	t.Run("valid server details", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &common.ServerModeConfig{
			ServerDetails: common.ServerDetailsConfig{
				Port:                 8080,
				CertificateDirectory: []string{tempDir},
			},
		}
		assert.NoError(t, validateServerDetails(config))
	})

	t.Run("missing port", func(t *testing.T) {
		config := &common.ServerModeConfig{
			ServerDetails: common.ServerDetailsConfig{
				CertificateDirectory: []string{t.TempDir()},
			},
		}
		assert.Error(t, validateServerDetails(config))
	})

	t.Run("missing certificate directory", func(t *testing.T) {
		config := &common.ServerModeConfig{
			ServerDetails: common.ServerDetailsConfig{
				Port: 8080,
			},
		}
		assert.Error(t, validateServerDetails(config))
	})

	t.Run("non-existent certificate directory", func(t *testing.T) {
		config := &common.ServerModeConfig{
			ServerDetails: common.ServerDetailsConfig{
				Port:                 8080,
				CertificateDirectory: []string{"nonexistentdir"},
			},
		}
		assert.Error(t, validateServerDetails(config))
	})
}

func TestValidateAgeKeys(t *testing.T) {
	_, publicKey := common.NewAgeTestKey(t)

	t.Run("valid age keys", func(t *testing.T) {
		config := &common.ServerModeConfig{
			PublicAgeKeys: []string{publicKey},
		}
		assert.NoError(t, validateAgeKeys(config))
	})

	t.Run("no age keys", func(t *testing.T) {
		config := &common.ServerModeConfig{}
		assert.Error(t, validateAgeKeys(config))
	})

	t.Run("invalid age key", func(t *testing.T) {
		config := &common.ServerModeConfig{
			PublicAgeKeys: []string{"invalid-key"},
		}
		assert.Error(t, validateAgeKeys(config))
	})
}
