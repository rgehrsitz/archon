package id

import (
	"strings"
	"testing"
	"time"
)

func TestNewV7(t *testing.T) {
	// Generate multiple UUIDs
	uuid1 := NewV7()
	uuid2 := NewV7()
	
	// Test basic format
	if len(uuid1) != 36 {
		t.Errorf("UUID length should be 36, got %d", len(uuid1))
	}
	
	if len(uuid2) != 36 {
		t.Errorf("UUID length should be 36, got %d", len(uuid2))
	}
	
	// Test uniqueness
	if uuid1 == uuid2 {
		t.Error("UUIDs should be unique")
	}
	
	// Test format (8-4-4-4-12)
	parts1 := strings.Split(uuid1, "-")
	if len(parts1) != 5 {
		t.Errorf("UUID should have 5 parts separated by hyphens, got %d", len(parts1))
	}
	
	expectedLengths := []int{8, 4, 4, 4, 12}
	for i, part := range parts1 {
		if len(part) != expectedLengths[i] {
			t.Errorf("Part %d should have length %d, got %d", i, expectedLengths[i], len(part))
		}
	}
	
	// Test that they are valid UUIDs
	if !IsValid(uuid1) {
		t.Errorf("Generated UUID should be valid: %s", uuid1)
	}
	
	if !IsValid(uuid2) {
		t.Errorf("Generated UUID should be valid: %s", uuid2)
	}
}

func TestUUIDV7TimeOrdering(t *testing.T) {
	// Generate UUIDs with small time gap
	uuid1 := NewV7()
	time.Sleep(1 * time.Millisecond)
	uuid2 := NewV7()
	
	// UUIDs should be lexicographically ordered by time
	if uuid1 >= uuid2 {
		t.Errorf("Later UUID should be lexicographically larger: %s >= %s", uuid1, uuid2)
	}
	
	// Extract timestamps
	ts1, err1 := ExtractTimestamp(uuid1)
	ts2, err2 := ExtractTimestamp(uuid2)
	
	if err1 != nil {
		t.Errorf("Failed to extract timestamp from uuid1: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Failed to extract timestamp from uuid2: %v", err2)
	}
	
	if ts1 >= ts2 {
		t.Errorf("Later timestamp should be larger: %d >= %d", ts1, ts2)
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		uuid  string
		valid bool
	}{
		{"Valid UUID", "01234567-89ab-cdef-0123-456789abcdef", true},
		{"Valid UUID uppercase", "01234567-89AB-CDEF-0123-456789ABCDEF", true},
		{"Empty string", "", false},
		{"Too short", "01234567-89ab-cdef-0123-456789abcde", false},
		{"Too long", "01234567-89ab-cdef-0123-456789abcdef0", false},
		{"Missing hyphens", "0123456789abcdef0123456789abcdef", false},
		{"Wrong hyphen positions", "012345678-9ab-cdef-0123-456789abcdef", false},
		{"Invalid hex characters", "01234567-89ab-ghij-0123-456789abcdef", false},
		{"Nil UUID", "00000000-0000-0000-0000-000000000000", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValid(tt.uuid)
			if result != tt.valid {
				t.Errorf("IsValid(%s) = %v, want %v", tt.uuid, result, tt.valid)
			}
		})
	}
}

func TestExtractTimestamp(t *testing.T) {
	// Test with generated UUID
	beforeTime := time.Now().UnixMilli()
	uuid := NewV7()
	afterTime := time.Now().UnixMilli()
	
	timestamp, err := ExtractTimestamp(uuid)
	if err != nil {
		t.Errorf("Failed to extract timestamp from generated UUID: %v", err)
	}
	
	// Timestamp should be within the time window
	if timestamp < beforeTime || timestamp > afterTime {
		t.Errorf("Extracted timestamp %d should be between %d and %d", timestamp, beforeTime, afterTime)
	}
	
	// Test with invalid UUID
	_, err = ExtractTimestamp("invalid-uuid")
	if err == nil {
		t.Error("Should fail to extract timestamp from invalid UUID")
	}
	
	// Test with valid but non-time-based UUID
	_, err = ExtractTimestamp("00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Errorf("Should be able to extract timestamp from nil UUID: %v", err)
	}
}

func TestUUIDUniqueness(t *testing.T) {
	// Generate many UUIDs quickly and ensure they're unique
	const numUUIDs = 10000
	uuids := make(map[string]bool)
	
	for range numUUIDs {
		uuid := NewV7()
		if uuids[uuid] {
			t.Errorf("Duplicate UUID generated: %s", uuid)
		}
		uuids[uuid] = true
	}
	
	if len(uuids) != numUUIDs {
		t.Errorf("Expected %d unique UUIDs, got %d", numUUIDs, len(uuids))
	}
}

func TestUUIDVersionBits(t *testing.T) {
	uuid := NewV7()
	
	// Remove hyphens and convert to bytes for bit checking
	cleanUUID := strings.ReplaceAll(uuid, "-", "")
	
	// Check version bits (should be 0111 = 7 in the 13th hex character)
	versionChar := cleanUUID[12] // 13th character (0-indexed)
	
	// Version should be 7, so the character should be '7'
	if versionChar != '7' {
		t.Errorf("Version bits should be 7, got character '%c' in position 12 of %s", versionChar, cleanUUID)
	}
	
	// Check variant bits (should be 10xx in the 17th hex character)
	variantChar := cleanUUID[16] // 17th character (0-indexed)
	
	// Variant bits should be 10xx, which means the hex digit should be 8, 9, A, or B
	validVariantChars := map[rune]bool{'8': true, '9': true, 'A': true, 'B': true, 'a': true, 'b': true}
	if !validVariantChars[rune(variantChar)] {
		t.Errorf("Variant bits should be 10xx (hex 8,9,A,B), got '%c' in position 16 of %s", variantChar, cleanUUID)
	}
}

func BenchmarkNewV7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewV7()
	}
}

func BenchmarkIsValid(b *testing.B) {
	uuid := NewV7()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		IsValid(uuid)
	}
}

func BenchmarkExtractTimestamp(b *testing.B) {
	uuid := NewV7()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ExtractTimestamp(uuid)
	}
}