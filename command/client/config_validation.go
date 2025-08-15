package client

import (
	"fmt"
	"go-certdist/common"
	"strings"

	"filippo.io/age"
	"github.com/rs/zerolog/log"
)

func validateConfig(config *common.ClientModeConfig) error {
	// Validate that the server connection is configured
	if err := validateConnectionDetails(config); err != nil {
		return err
	}

	// Validate age public and private key
	if err := validateAgeKeys(config); err != nil {
		return err
	}

	// Validate that at least one certificate config exists with a domain and a directory
	if err := validateCertificates(config); err != nil {
		return err
	}
	return nil
}

func validateConnectionDetails(config *common.ClientModeConfig) error {
	if config.ConnectionDetails.Server == "" {
		return fmt.Errorf("connection.server is not configured")
	}

	if strings.HasPrefix(config.ConnectionDetails.Server, "http://") {
		log.Warn().Str("server", config.ConnectionDetails.Server).Msg("Using insecure HTTP connection")
	} else if !strings.HasPrefix(config.ConnectionDetails.Server, "https://") {
		config.ConnectionDetails.Server = "https://" + config.ConnectionDetails.Server
		log.Info().Str("server", config.ConnectionDetails.Server).Msg("Unknown protocol, assuming HTTPs")
	}

	if strings.HasSuffix(config.ConnectionDetails.Server, "/") {
		config.ConnectionDetails.Server = config.ConnectionDetails.Server[:len(config.ConnectionDetails.Server)-1]
	}
	return nil
}

func validateAgeKeys(config *common.ClientModeConfig) error {
	if config.AgeKey.PrivateKey == "" {
		return fmt.Errorf("age_key.private_key is not configured")
	}

	identity, err := age.ParseX25519Identity(config.AgeKey.PrivateKey)
	if err != nil {
		return fmt.Errorf("invalid age_key.private_key: %w", err)
	}

	if config.AgeKey.PublicKey == "" { // auto-create public key if not provided
		config.AgeKey.PublicKey = identity.Recipient().String()
	}
	if identity.Recipient().String() != config.AgeKey.PublicKey {
		return fmt.Errorf("Public key does not correspond to the private key")
	}
	return nil
}

func validateCertificates(config *common.ClientModeConfig) error {
	if len(config.Certificate) == 0 {
		return fmt.Errorf("at least one certificate must be configured")
	}

	for i, certConfig := range config.Certificate {
		if certConfig.Domain == "" {
			return fmt.Errorf("certificate %d: domain is not configured", i)
		}
		if certConfig.Directory == "" {
			return fmt.Errorf("certificate %d: directory is not configured for domain %s", i, certConfig.Domain)
		}
	}
	return nil
}
