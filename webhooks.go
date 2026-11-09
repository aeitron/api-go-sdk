package aeitron

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidSignatureHeader = errors.New("aeitron: invalid signature header format (expected t=...,v1=...)")
	ErrTimestampExpired       = errors.New("aeitron: webhook timestamp outside acceptable tolerance (possible replay attack)")
	ErrSignatureMismatch      = errors.New("aeitron: webhook signature verification failed")
)

// WebhookVerifier performs cryptographic verification of Aeitron webhook events.
type WebhookVerifier struct{}

// NewWebhookVerifier creates a new WebhookVerifier instance.
func NewWebhookVerifier() *WebhookVerifier {
	return &WebhookVerifier{}
}

// VerifySignature validates the Aeitron-Signature header against the raw payload.
// header format: t=1730000000,v1=9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
func (w *WebhookVerifier) VerifySignature(payload []byte, signatureHeader, secret string, tolerance time.Duration) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return errors.New("aeitron: webhook secret cannot be empty")
	}

	parts := strings.Split(signatureHeader, ",")
	var timestampStr string
	var signatures []string

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := kv[0], kv[1]
		if key == "t" {
			timestampStr = val
		} else if key == "v1" {
			signatures = append(signatures, val)
		}
	}

	if timestampStr == "" || len(signatures) == 0 {
		return ErrInvalidSignatureHeader
	}

	tsUnix, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return ErrInvalidSignatureHeader
	}

	eventTime := time.Unix(tsUnix, 0)
	if tolerance > 0 {
		diff := time.Since(eventTime)
		if diff < -tolerance || diff > tolerance {
			return fmt.Errorf("%w: event age %v exceeds tolerance %v", ErrTimestampExpired, diff, tolerance)
		}
	}

	// Signed payload is "${timestamp}.${payload}"
	signedPayload := fmt.Sprintf("%s.%s", timestampStr, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expectedSignature)) {
			return nil
		}
	}

	return ErrSignatureMismatch
}

