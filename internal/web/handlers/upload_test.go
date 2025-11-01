package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single tag",
			input:    "vacation",
			expected: []string{"vacation"},
		},
		{
			name:     "multiple tags",
			input:    "vacation,sunset,beach",
			expected: []string{"vacation", "sunset", "beach"},
		},
		{
			name:     "tags with spaces",
			input:    "vacation, sunset , beach",
			expected: []string{"vacation", "sunset", "beach"},
		},
		{
			name:     "tags with empty values",
			input:    "vacation,,beach",
			expected: []string{"vacation", "beach"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSupportedImageType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"image/jpeg", true},
		{"image/jpg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"image/bmp", false},
		{"text/plain", false},
		{"application/pdf", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := isSupportedImageType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
