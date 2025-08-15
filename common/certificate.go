package common

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// FileType is an enum for the different types of files we can parse.
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypePublicCertificate
	FileTypePrivateKey
)

// String makes FileType implement the Stringer interface for easy printing.
func (ft FileType) String() string {
	switch ft {
	case FileTypePublicCertificate:
		return "Public Certificate"
	case FileTypePrivateKey:
		return "Private Key"
	default:
		return "Unknown"
	}
}

// CertificateInfo holds the extracted details of a certificate.
type CertificateInfo struct {
	Domains    []string
	Expiration time.Time
	FilePath   string
	FileType   FileType // e.g., "Public Certificate", "Private Key"
}

type DirectoryCertificates struct {
	FilePath     string
	Certificates []*CertificateInfo
}

// parseCertificateFile reads a PEM-encoded file and extracts its details.
func parseCertificateFile(filePath string) (*CertificateInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	info := &CertificateInfo{
		FilePath: filePath,
		FileType: FileTypeUnknown,
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// We just check the first block for simplicity
	if block.Type == "CERTIFICATE" {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		info.FileType = FileTypePublicCertificate
		info.Domains = cert.DNSNames
		info.Expiration = cert.NotAfter
	} else if strings.Contains(block.Type, "PRIVATE KEY") {
		info.FileType = FileTypePrivateKey
		// Private keys don't have domain or expiration info in them.
	} else {
		return nil, fmt.Errorf("unhandled PEM block type: %s", block.Type)
	}

	if len(data) > 0 {
		// Ignore the rest of the file, fullchain.pem does have the chain.pem appended, which we don't care about
	}

	return info, nil
}

// LoadCertificates loads given directories and parse their certificates, does not recurse
func LoadCertificates(directories []string) []DirectoryCertificates {
	var result []DirectoryCertificates
	for _, dir := range directories {
		log.Info().Str("directory", dir).Msg("Parsing certificate directory")
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Error().Err(err).Str("directory", dir).Msg("Failed to read certificate directory")
			continue
		}

		var certificates []*CertificateInfo
		for _, file := range files {
			if file.IsDir() {
				continue // Skip directories
			}
			path := filepath.Join(dir, file.Name())
			info, err := parseCertificateFile(path)
			if err != nil {
				log.Warn().Err(err).Str("file", path).Msg("Failed to parse certificate file")
				continue
			}
			if info != nil {
				certificates = append(certificates, info)
			}
		}
		result = append(result, DirectoryCertificates{
			FilePath:     dir,
			Certificates: certificates,
		})
	}
	return result
}

// DebugPrintCertificates logs the certificates found in the given directories,
// including their expiration dates, types, and domains. This is meant for
// debugging purposes only.
func DebugPrintCertificates(certificates []DirectoryCertificates) {
	for _, dir := range certificates {
		log.Info().Str("directory", dir.FilePath).Msg("Certificates found in directory")
		for _, cert := range dir.Certificates {
			log.Info().
				Str("expiration", cert.Expiration.Format(time.RFC3339)).
				Str("file", cert.FilePath).
				Str("type", cert.FileType.String()).
				Str("domains", strings.Join(cert.Domains, ", ")).
				Msg("Certificate:")
		}
	}
}

func FindCertificate(certificateDirectories []DirectoryCertificates, domain string) []*CertificateInfo {
	var result []*CertificateInfo
	for _, dir := range certificateDirectories {
		found := false
		for _, cert := range dir.Certificates {
			for _, certDomain := range cert.Domains {
				// TODO this is a hack
				if strings.Contains(certDomain, domain) || strings.Contains(certDomain, "*."+domain) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			for _, cert := range dir.Certificates {
				result = append(result, cert)
			}
		}
	}
	return result
}
