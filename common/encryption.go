package common

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"
)

// EncryptAndZipCertificates takes a list of certificate info, zips the corresponding files,
// and encrypts the zip archive using the provided age public key.
func EncryptAndZipCertificates(certificates []*CertificateInfo, publicKey string) ([]byte, error) {
	// 1. Create a buffer to write our zip archive to.
	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)

	// 2. Add files to the zip archive.
	for _, certInfo := range certificates {
		fileData, err := os.ReadFile(certInfo.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", certInfo.FilePath, err)
		}

		zipFile, err := zipWriter.Create(filepath.Base(certInfo.FilePath))
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry for %s: %w", certInfo.FilePath, err)
		}

		_, err = zipFile.Write(fileData)
		if err != nil {
			return nil, fmt.Errorf("failed to write file content to zip for %s: %w", certInfo.FilePath, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	// 3. Encrypt the zip buffer with the age public key.
	recipient, err := age.ParseX25519Recipient(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse age public key: %w", err)
	}

	encryptedBuf := new(bytes.Buffer)
	w, err := age.Encrypt(encryptedBuf, recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption writer: %w", err)
	}

	if _, err := io.Copy(w, zipBuf); err != nil {
		return nil, fmt.Errorf("failed to write to encryption writer: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encryption writer: %w", err)
	}

	return encryptedBuf.Bytes(), nil
}

// GenerateAndPrintKeyPair creates a new X25519 key pair and prints the private and public keys.
func GenerateAndPrintKeyPair() error {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	privateKey := identity.String()
	publicKey := identity.Recipient().String()

	fmt.Printf("# Randomly generated age key pair, can also be generated with 'age-keygen'\n")
	fmt.Printf("AGE_PRIVATE_KEY=%s\n", privateKey)
	fmt.Printf("AGE_PUBLIC_KEY=%s\n", publicKey)

	return nil
}
