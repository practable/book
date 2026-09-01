package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	HeaderTimestamp   = "X-Book-Timestamp"
	HeaderDeliveryID  = "X-Book-Delivery-ID"
	HeaderDirection   = "X-Book-Direction"
	HeaderSignature   = "X-Book-Signature"
	DirectionCommand  = "book-to-runner"
	DirectionCallback = "runner-to-book"
)

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrExpiredSignature = errors.New("webhook signature timestamp outside tolerance")
)

type Headers struct {
	Timestamp  string
	DeliveryID string
	Direction  string
	Signature  string
}

// ParseSecret accepts the documented 32-byte hex or unpadded base64url form.
func ParseSecret(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	var decoded []byte
	var err error
	if len(value) == 64 {
		decoded, err = hex.DecodeString(value)
	} else {
		decoded, err = base64.RawURLEncoding.DecodeString(value)
	}
	if err != nil || len(decoded) != 32 {
		return nil, errors.New("webhook secret must encode exactly 32 random bytes as hex or unpadded base64url")
	}
	return decoded, nil
}

func signedBytes(direction, timestamp, deliveryID string, body []byte) []byte {
	prefix := direction + "." + timestamp + "." + deliveryID + "."
	return append([]byte(prefix), body...)
}

func Sign(secret []byte, direction, deliveryID string, at time.Time, body []byte) (Headers, error) {
	if len(secret) != 32 {
		return Headers{}, errors.New("webhook secret must be 32 bytes")
	}
	if direction != DirectionCommand && direction != DirectionCallback {
		return Headers{}, errors.New("invalid webhook direction")
	}
	if strings.TrimSpace(deliveryID) == "" {
		return Headers{}, errors.New("delivery ID is required")
	}
	timestamp := strconv.FormatInt(at.UTC().Unix(), 10)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(signedBytes(direction, timestamp, deliveryID, body))
	return Headers{Timestamp: timestamp, DeliveryID: deliveryID, Direction: direction, Signature: "v1=" + hex.EncodeToString(mac.Sum(nil))}, nil
}

func Verify(secret []byte, expectedDirection string, headers Headers, now time.Time, tolerance time.Duration, body []byte) error {
	if len(secret) != 32 || headers.Direction != expectedDirection || strings.TrimSpace(headers.DeliveryID) == "" {
		return ErrInvalidSignature
	}
	seconds, err := strconv.ParseInt(headers.Timestamp, 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}
	signedAt := time.Unix(seconds, 0)
	delta := now.UTC().Sub(signedAt)
	if delta < 0 {
		delta = -delta
	}
	if tolerance <= 0 || delta > tolerance {
		return ErrExpiredSignature
	}
	provided, err := hex.DecodeString(strings.TrimPrefix(headers.Signature, "v1="))
	if err != nil || !strings.HasPrefix(headers.Signature, "v1=") {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(signedBytes(headers.Direction, headers.Timestamp, headers.DeliveryID, body))
	if !hmac.Equal(provided, mac.Sum(nil)) {
		return ErrInvalidSignature
	}
	return nil
}

func (h Headers) String() string {
	return fmt.Sprintf("%s %s %s", h.Direction, h.Timestamp, h.DeliveryID)
}
