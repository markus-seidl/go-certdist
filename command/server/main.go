package server

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"go-certdist/common"
	"io"
	"math/rand"
	"net/http"

	"github.com/rs/zerolog/log"
)

func StartServer(config common.ServerModeConfig) {
	if err := validateConfig(&config); err != nil {
		log.Fatal().Err(err).Msg("Server configuration validation failed")
	}

	http.HandleFunc(common.CertificateRequestEndpoint, handleCertificateRequest(config))
	http.HandleFunc(common.HealthEndpoint, handleHealthCheck)

	addr := fmt.Sprintf("%s:%d", config.ServerDetails.ListenAddress, config.ServerDetails.Port)
	log.Info().Str("address", addr).Msg("Starting server")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "OK")
}

func handleCertificateRequest(config common.ServerModeConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate a random request ID
		reqID := fmt.Sprintf("%x", rand.Uint32())
		logCtx := log.With().Str(common.LogKeyRequestId, reqID).Logger()
		logCtx.Info().Str("remoteAddr", r.RemoteAddr).Msg("Received request from IP")

		if r.Method != http.MethodPost {
			logCtx.Warn().Msg("Only POST method is allowed")
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read request body
		var req common.CertificateRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		logCtx.Debug().RawJSON("request_body", body).Msg("Received request body")
		if err := json.Unmarshal(body, &req); err != nil {
			logCtx.Error().Err(err).Msg("Failed to unmarshal request body")
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		logCtx.Info().
			Str("domain", req.Domain).
			Str("age_public_key", req.AgePublicKey).
			Time("expiration", req.Expiration).
			Msg("Received certificate request")

		// (Re-)Load certificates
		directoryCertificates := common.LoadCertificates(config.ServerDetails.CertificateDirectory)
		common.DebugPrintCertificates(directoryCertificates)
		foundCerts := common.FindCertificate(directoryCertificates, req.Domain)
		if len(foundCerts) == 0 {
			logCtx.Info().Str("domain", req.Domain).Msg("Certificate not found for domain")
			http.Error(w, "Certificate not found", http.StatusNotFound)
			return
		}

		// Validate expiration date
		if !req.Expiration.IsZero() {
			serverCertExpiration := foundCerts[0].Expiration
			if !serverCertExpiration.After(req.Expiration) {
				logCtx.Info().Msg("Client certificate is up to date. No action needed.")
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		logCtx.Info().Msg("Sending certificate")

		if err := validateAgePublicKey(config, req.AgePublicKey); err != nil {
			logCtx.Warn().Str("age_public_key", req.AgePublicKey).Msg("Public key not whitelisted")
			http.Error(w, "Public key not authorized", http.StatusForbidden)
			return
		}

		// Encrypt the certificates
		encryptedData, err := common.EncryptAndZipCertificates(foundCerts, req.AgePublicKey)
		if err != nil {
			logCtx.Error().Err(err).Msg("Failed to encrypt certificates")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		_, err = w.Write(encryptedData)
		if err != nil {
			logCtx.Error().Err(err).Msg("Failed to write response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func validateAgePublicKey(config common.ServerModeConfig, reqPublicKey string) error {
	// Secure comparison, always compare everything

	found := false
	for _, allowedKey := range config.PublicAgeKeys {
		if subtle.ConstantTimeCompare([]byte(allowedKey), []byte(reqPublicKey)) == 1 {
			found = true
			// no break, constant time comparisons
		}
	}

	if found {
		return nil
	}

	return fmt.Errorf("public key not whitelisted")
}
