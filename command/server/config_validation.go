package server

import (
	"fmt"
	"go-certdist/common"
	"os"

	"filippo.io/age"
)

func validateConfig(config *common.ServerModeConfig) error {
	// Validate that the server connection is configured
	if err := validateServerDetails(config); err != nil {
		return err
	}

	// Validate age public and private key
	if err := validateAgeKeys(config); err != nil {
		return err
	}
	return nil
}

func validateServerDetails(config *common.ServerModeConfig) error {
	if len(config.ServerDetails.ListenAddress) == 0 {
		config.ServerDetails.ListenAddress = "127.0.0.1"
	}

	if config.ServerDetails.Port == 0 {
		return fmt.Errorf("server.port is not configured")
	}
	if len(config.ServerDetails.CertificateDirectory) == 0 {
		return fmt.Errorf("at least one server.certificate_directories must be configured")
	}
	for _, dir := range config.ServerDetails.CertificateDirectory {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("configured certificate directory does not exist: %s", dir)
		}
	}
	return nil
}

func validateAgeKeys(config *common.ServerModeConfig) error {
	if len(config.PublicAgeKeys) == 0 {
		return fmt.Errorf("at least one public_age_key must be configured")
	}
	for _, key := range config.PublicAgeKeys {
		if _, err := age.ParseX25519Recipient(key); err != nil {
			return fmt.Errorf("invalid public_age_key configured: %s", key)
		}
	}
	return nil
}
