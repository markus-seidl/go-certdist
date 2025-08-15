package common

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptAndZipCertificates(t *testing.T) {
	tempDir := t.TempDir()

	// Create some dummy files to encrypt
	file1Path := filepath.Join(tempDir, "file1.txt")
	file1Content := []byte("this is file 1")
	require.NoError(t, os.WriteFile(file1Path, file1Content, 0644))

	file2Path := filepath.Join(tempDir, "file2.txt")
	file2Content := []byte("this is file 2")
	require.NoError(t, os.WriteFile(file2Path, file2Content, 0644))

	certInfos := []*CertificateInfo{
		{FilePath: file1Path},
		{FilePath: file2Path},
	}

	privateKey, publicKey := NewAgeTestKey(t)

	t.Run("Successful encryption and zipping", func(t *testing.T) {
		encryptedData, err := EncryptAndZipCertificates(certInfos, publicKey)
		require.NoError(t, err)
		assert.NotEmpty(t, encryptedData)

		// Decrypt the data to verify
		identity, err := age.ParseX25519Identity(privateKey)
		require.NoError(t, err)

		decryptor, err := age.Decrypt(bytes.NewReader(encryptedData), identity)
		require.NoError(t, err)

		decryptedZip, err := io.ReadAll(decryptor)
		require.NoError(t, err)

		// Unzip and check contents
		zipReader, err := zip.NewReader(bytes.NewReader(decryptedZip), int64(len(decryptedZip)))
		require.NoError(t, err)

		assert.Len(t, zipReader.File, 2)

		foundFiles := make(map[string][]byte)
		for _, f := range zipReader.File {
			rc, err := f.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			foundFiles[f.Name] = content
		}

		assert.Equal(t, file1Content, foundFiles["file1.txt"])
		assert.Equal(t, file2Content, foundFiles["file2.txt"])
	})

	t.Run("Invalid public key", func(t *testing.T) {
		_, err := EncryptAndZipCertificates(certInfos, "invalid-pub-key")
		assert.Error(t, err)
	})

	t.Run("Non-existent input file", func(t *testing.T) {
		badCertInfos := []*CertificateInfo{
			{FilePath: "non-existent-file.txt"},
		}
		_, err := EncryptAndZipCertificates(badCertInfos, publicKey)
		assert.Error(t, err)
	})
}
