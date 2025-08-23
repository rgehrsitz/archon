package id

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// NewV7 generates a UUIDv7 (time-sortable) as specified in ADR-001
// Format: 48-bit timestamp (ms) + 12-bit random + 2-bit version + 62-bit random
func NewV7() string {
	var uuid [16]byte
	
	// Get current timestamp in milliseconds
	now := time.Now().UnixMilli()
	
	// Set timestamp in first 48 bits (6 bytes)
	binary.BigEndian.PutUint64(uuid[:8], uint64(now)<<16)
	
	// Fill remaining 10 bytes with random data
	if _, err := rand.Read(uuid[6:]); err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	
	// Set version bits (4 bits): version 7
	uuid[6] = (uuid[6] & 0x0f) | 0x70
	
	// Set variant bits (2 bits): RFC 4122 variant (10)
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(uuid[0:4]),
		binary.BigEndian.Uint16(uuid[4:6]),
		binary.BigEndian.Uint16(uuid[6:8]),
		binary.BigEndian.Uint16(uuid[8:10]),
		uuid[10:16])
}

// IsValid checks if a string is a valid UUID format
func IsValid(id string) bool {
	if len(id) != 36 {
		return false
	}
	
	// Check format: 8-4-4-4-12 characters with hyphens
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		return false
	}
	
	// Check if all other characters are hex
	for i, c := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	
	return true
}

// ExtractTimestamp extracts the timestamp from a UUIDv7 (returns Unix milliseconds)
func ExtractTimestamp(id string) (int64, error) {
	if !IsValid(id) {
		return 0, fmt.Errorf("invalid UUID format")
	}
	
	// Remove hyphens and take first 12 hex characters (48 bits = 6 bytes)
	cleanID := id[:8] + id[9:13] + id[14:18] // First 8 + next 4 + next 4 = 16 chars for 8 bytes
	timestampHex := cleanID[:12] // First 12 hex chars = 48 bits = 6 bytes
	
	// Parse the timestamp hex string
	var timestamp uint64
	if _, err := fmt.Sscanf(timestampHex, "%012x", &timestamp); err != nil {
		return 0, fmt.Errorf("failed to parse timestamp: %v", err)
	}
	
	return int64(timestamp), nil
}
