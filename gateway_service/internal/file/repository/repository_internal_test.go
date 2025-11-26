package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractObjectNameFromURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid URL",
			url:     "http://minio:9000/bucket/object-name.jpg",
			want:    "object-name.jpg",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			url:     "invalid-url",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty URL",
			url:     "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractObjectNameFromURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
