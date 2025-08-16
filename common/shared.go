package common

import "time"

const CertificateRequestEndpoint = "/api/v1/certificate-request"

//
// Server
//

type ServerDetailsConfig struct {
	Port                 int32    `yaml:"port"`
	ListenAddress        string   `yaml:"listen_address,omitempty"`
	CertificateDirectory []string `yaml:"certificate_directories"`
}

// ServerModeConfig defines the structure for the server configuration.
type ServerModeConfig struct {
	ServerDetails ServerDetailsConfig `yaml:"server"`
	PublicAgeKeys []string            `yaml:"public_age_keys"`
}

//
// Client
//

type CertificateConfig struct {
	Domain        string   `yaml:"domain"`
	Directory     string   `yaml:"directory"`
	RenewCommands []string `yaml:"renew_commands,omitempty"`
}

type ClientConnectionConfig struct {
	Server string `yaml:"server"`
}

type AgeKeyConfig struct {
	PublicKey  string `yaml:"public_key,omitempty"`
	PrivateKey string `yaml:"private_key"`
}

// ClientModeConfig defines the structure for the client configuration.
type ClientModeConfig struct {
	ConnectionDetails ClientConnectionConfig `yaml:"connection"`
	Certificate       []CertificateConfig    `yaml:"certificate"`
	AgeKey            AgeKeyConfig           `yaml:"age_key"`
}

//
// API
//

// CertificateRequest defines the structure for the certificate request JSON body.
type CertificateRequest struct {
	Domain       string    `json:"domain"`
	AgePublicKey string    `json:"age_public_key"`
	Expiration   time.Time `json:"expiration"`
}
