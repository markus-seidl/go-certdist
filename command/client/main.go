package client

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"go-certdist/common"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/rs/zerolog/log"
)

func ExecuteClient(config common.ClientModeConfig) {
	if err := validateConfig(&config); err != nil {
		log.Fatal().Err(err).Msg("Client configuration validation failed")
	}

	for {
		for _, certConfig := range config.Certificate {
			log.Info().Msg("==================================================================")
			log.Info().Str("server", config.ConnectionDetails.Server).Str("domain", certConfig.Domain).Msg("Requesting certificate from server")
			if err := processCertificateRequest(config, certConfig); err != nil {
				log.Error().Err(err).Str("domain", certConfig.Domain).Msg("Failed to process certificate request")
			}
		}

		if config.Interval <= 0 {
			log.Info().Msg("Interval not configured, exiting after single execution")
			break
		}

		waitDuration := time.Duration(config.Interval) * time.Hour
		log.Info().Int("hours", config.Interval).Msg("Waiting until next execution")
		time.Sleep(waitDuration)
	}
}

func processCertificateRequest(config common.ClientModeConfig, certConfig common.CertificateConfig) error {
	// Check for existing certificate and its expiration date
	var expirationDate time.Time
	if _, err := os.Stat(certConfig.Directory); !os.IsNotExist(err) {
		log.Info().Str("directory", certConfig.Directory).Msg("Checking existing certificates")
		existingCerts := common.LoadCertificates([]string{certConfig.Directory})
		foundCerts := common.FindCertificate(existingCerts, certConfig.Domain)
		if len(foundCerts) > 0 {
			expirationDate = foundCerts[0].Expiration
			log.Info().Time("expiration", expirationDate).Msg("Found existing certificate expiration date")
		}
	}

	resp, err := sendRequestToServer(config, certConfig, expirationDate)
	if err != nil {
		return fmt.Errorf("failed to send certificate request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusNotModified {
		log.Info().Str("domain", certConfig.Domain).Msg("Certificate is up-to-date")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get certificate with status %d: %s", resp.StatusCode, string(body))
	}

	encryptedData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 3. Decrypt the data
	decryptedData, err := decryptWithAge(encryptedData, config.AgeKey.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	// 4. Unzip the decrypted data
	if err := unzip(decryptedData, certConfig.Directory); err != nil {
		return fmt.Errorf("failed to unzip data: %w", err)
	}

	log.Info().Str("directory", certConfig.Directory).Str("domain", certConfig.Domain).Msg("Successfully downloaded and extracted certificates")

	// 5. Execute renew commands
	if err := executeRenewCommands(certConfig.RenewCommands); err != nil {
		return fmt.Errorf("failed to execute one or more renew commands: %w", err)
	}

	return nil
}

func sendRequestToServer(config common.ClientModeConfig, certConfig common.CertificateConfig, expirationDate time.Time) (*http.Response, error) {
	// 1. Prepare the request body
	reqBody := common.CertificateRequest{
		Domain:       certConfig.Domain,
		AgePublicKey: config.AgeKey.PublicKey,
		Expiration:   expirationDate,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// 2. Send the POST request
	url := fmt.Sprintf("%s%s", config.ConnectionDetails.Server, common.CertificateRequestEndpoint)
	log.Debug().Str("url", url).Msg("Sending request to server")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func executeRenewCommands(commands []string) error {
	for _, command := range commands {
		log.Info().Str("command", command).Msg("Executing renew command")
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute command '%s': %w", command, err)
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			log.Info().Str("command", command).Str("line", line).Msg("Command output")
		}
		log.Info().Str("command", command).Msg("Successfully executed renew command")
	}
	return nil
}

func decryptWithAge(data []byte, privateKey string) ([]byte, error) {
	identity, err := age.ParseX25519Identity(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse age private key: %w", err)
	}

	r := bytes.NewReader(data)
	decryptor, err := age.Decrypt(r, identity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	out := &bytes.Buffer{}
	if _, err := io.Copy(out, decryptor); err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return out.Bytes(), nil
}

func unzip(data []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			continue // don't create empty directories
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer func(outFile *os.File) {
			err := outFile.Close()
			if err != nil {
				log.Warn().Err(err).Str("file", fpath).Msg("Failed to close file on disk")
			}
		}(outFile)

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func(rc io.ReadCloser) {
			err := rc.Close()
			if err != nil {
				log.Warn().Err(err).Str("file", fpath).Msg("Failed to close file from zip archive")
			}
		}(rc)

		_, err = io.Copy(outFile, rc)

		if err != nil {
			return err
		}
	}
	return nil
}
