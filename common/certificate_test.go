package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary test certificate and private key
func createTestCert(t *testing.T, dir, domain string) (certPath, keyPath string) {
	t.Helper()

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1337),
		Subject: pkix.Name{
			Organization: []string{"Test Corp"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 30),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	// Create certificate file
	certFile, err := os.Create(filepath.Join(dir, "cert.pem"))
	require.NoError(t, err)
	defer certFile.Close()
	require.NoError(t, pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))

	// Create private key file
	keyFile, err := os.Create(filepath.Join(dir, "privkey.pem"))
	require.NoError(t, err)
	defer keyFile.Close()
	keyBytes, err := x509.MarshalECPrivateKey(privKey)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}))

	return certFile.Name(), keyFile.Name()
}

func TestParseCertificateFile(t *testing.T) {
	tempDir := t.TempDir()
	domain := "example.com"
	certPath, keyPath := createTestCert(t, tempDir, domain)

	t.Run("Valid Public Certificate", func(t *testing.T) {
		info, err := parseCertificateFile(certPath)
		require.NoError(t, err)
		assert.Equal(t, FileTypePublicCertificate, info.FileType)
		assert.Contains(t, info.Domains, domain)
		assert.WithinDuration(t, time.Now().Add(time.Hour*24*30), info.Expiration, time.Second*5)
	})

	t.Run("Valid Private Key", func(t *testing.T) {
		info, err := parseCertificateFile(keyPath)
		require.NoError(t, err)
		assert.Equal(t, FileTypePrivateKey, info.FileType)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		_, err := parseCertificateFile("nonexistent.pem")
		assert.Error(t, err)
	})

	t.Run("Invalid PEM file", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.pem")
		require.NoError(t, os.WriteFile(invalidFile, []byte("not a pem file"), 0644))
		_, err := parseCertificateFile(invalidFile)
		assert.Error(t, err)
	})
}

func TestLoadCertificates(t *testing.T) {
	tempDir := t.TempDir()
	createTestCert(t, tempDir, "example.com")

	t.Run("Valid directory", func(t *testing.T) {
		dirs := []string{tempDir}
		certs := LoadCertificates(dirs)
		require.Len(t, certs, 1)
		assert.Len(t, certs[0].Certificates, 2)
	})

	t.Run("Non-existent directory", func(t *testing.T) {
		dirs := []string{"nonexistentdir"}
		certs := LoadCertificates(dirs)
		assert.Len(t, certs, 0)
	})
}

func TestFindCertificate(t *testing.T) {
	domain := "example.com"
	certInfo := &CertificateInfo{
		Domains:  []string{domain},
		FileType: FileTypePublicCertificate,
	}
	directoryCerts := []DirectoryCertificates{
		{
			FilePath:     "/some/path",
			Certificates: []*CertificateInfo{certInfo},
		},
	}

	t.Run("Find existing certificate", func(t *testing.T) {
		found := FindCertificate(directoryCerts, domain)
		require.Len(t, found, 1)
		assert.Equal(t, certInfo, found[0])
	})

	t.Run("Find non-existent certificate", func(t *testing.T) {
		found := FindCertificate(directoryCerts, "notfound.com")
		assert.Len(t, found, 0)
	})
}
