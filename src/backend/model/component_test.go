// Package model provides the core data models for Archon.
package model

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	c := NewComponent("test-id", "Test Component", "device")
	
	if c.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", c.ID)
	}
	
	if c.Name != "Test Component" {
		t.Errorf("Expected Name 'Test Component', got '%s'", c.Name)
	}
	
	if c.Type != "device" {
		t.Errorf("Expected Type 'device', got '%s'", c.Type)
	}
	
	if c.Properties == nil {
		t.Error("Properties map should be initialized")
	}
	
	if c.Metadata == nil {
		t.Error("Metadata map should be initialized")
	}
}

func TestComponent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		comp    *Component
		wantErr bool
	}{
		{
			name: "Valid component",
			comp: &Component{
				ID:         "valid-id",
				Name:       "Valid Component",
				Type:       "device",
				Properties: map[string]interface{}{},
				Metadata:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "Missing ID",
			comp: &Component{
				Name:       "Invalid Component",
				Type:       "device",
				Properties: map[string]interface{}{},
				Metadata:   map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "Missing Name",
			comp: &Component{
				ID:         "test-id",
				Type:       "device",
				Properties: map[string]interface{}{},
				Metadata:   map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "Missing Type",
			comp: &Component{
				ID:         "test-id",
				Name:       "Test Component",
				Properties: map[string]interface{}{},
				Metadata:   map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid ID characters",
			comp: &Component{
				ID:         "invalid id with spaces",
				Name:       "Invalid Component",
				Type:       "device",
				Properties: map[string]interface{}{},
				Metadata:   map[string]string{},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		id       string
		expected bool
	}{
		{"valid-id", true},
		{"valid_id", true},
		{"valid123", true},
		{"VALID_ID", true},
		{"", false},
		{"invalid id", false},
		{"invalid.id", false},
		{"invalid@id", false},
		{"veryLongIDThatExceedsSixtyFourCharactersLimitAndShouldBeRejectedByValidation", false},
	}
	
	for _, test := range tests {
		result := isValidID(test.id)
		if result != test.expected {
			t.Errorf("isValidID(%q) = %v, want %v", test.id, result, test.expected)
		}
	}
}
