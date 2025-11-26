package utils

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTransformMinioURL(t *testing.T) {
	viper.Set("MINIO_ENDPOINT", "http://minio:9000")
	viper.Set("MINIO_PUBLIC_ENDPOINT", "http://localhost:9000")

	tests := []struct {
		name        string
		internalURL string
		expected    string
	}{
		{
			name:        "Empty URL",
			internalURL: "",
			expected:    "",
		},
		{
			name:        "Transform Internal to Public",
			internalURL: "http://minio:9000/bucket/file.jpg",
			expected:    "http://localhost:9000/bucket/file.jpg",
		},
		{
			name:        "No Minio Config",
			internalURL: "http://minio:9000/bucket/file.jpg",
			expected:    "http://minio:9000/bucket/file.jpg",
		},
	}

	for _, tt := range tests[:2] {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformMinioURL(tt.internalURL)
			assert.Equal(t, tt.expected, got)
		})
	}

	viper.Set("MINIO_ENDPOINT", "")
	viper.Set("MINIO_PUBLIC_ENDPOINT", "")
	t.Run(tests[2].name, func(t *testing.T) {
		got := TransformMinioURL(tests[2].internalURL)
		assert.Equal(t, tests[2].expected, got)
	})
}

func TestTransformShareURL(t *testing.T) {
	viper.Set("APP_BASE_URL", "http://example.com")

	tests := []struct {
		name      string
		shareUUID string
		expected  string
	}{
		{
			name:      "Empty UUID",
			shareUUID: "",
			expected:  "",
		},
		{
			name:      "Valid UUID",
			shareUUID: "uuid-123",
			expected:  "http://example.com/share/uuid-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformShareURL(tt.shareUUID)
			assert.Equal(t, tt.expected, got)
		})
	}
}
