package common

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func LoadServerConfig(path string) ServerModeConfig {
	var c ServerModeConfig
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Str("path", path).Msg("Failed to read config file")
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Fatal().Err(err).Msg("Whoops, something is wrong with your config!")
	}

	return c
}

func LoadClientConfig(path string) ClientModeConfig {
	var c ClientModeConfig
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Str("path", path).Msg("Failed to read config file")
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		log.Fatal().Err(err).Msg("Whoops, something is wrong with your config!")
	}

	return c
}

func WriteDummyServerConfig() {
	dummyConfig := ServerModeConfig{
		ServerDetails: ServerDetailsConfig{
			Port:                 8080,
			CertificateDirectory: []string{"/path/to/certificates"},
		},
		PublicAgeKeys: []string{"age1publickey"},
	}

	data, err := yaml.Marshal(&dummyConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal server config")
	}

	fmt.Println(string(data))
}

func WriteDummyClientConfig() {
	dummyConfig := ClientModeConfig{
		ConnectionDetails: ClientConnectionConfig{
			Server: "localhost:8080",
		},
		AgeKey: AgeKeyConfig{
			PrivateKey: "age1privatekey",
		},
		Certificate: []CertificateConfig{
			{
				Domain:        "example.com",
				Directory:     "/path/to/output/dir",
				RenewCommands: []string{"echo 'renewing certificate cmd 1'", "echo 'cmd 2'"},
			},
		},
	}

	data, err := yaml.Marshal(&dummyConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal client config")
	}

	fmt.Println(string(data))
}
