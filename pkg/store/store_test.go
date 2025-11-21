package store

import (
	"context"
	"strings"
	"testing"
)

func TestNewStoreAndUploadFileNoMinio(t *testing.T) {
    s := NewStore()
    if s == nil {
        t.Fatalf("NewStore returned nil")
    }

    // UploadFileToMinIO should return error when Minio is not initialized
    _, err := s.UploadFileToMinIO(context.Background(), "file.txt", strings.NewReader("data"), 4, "text/plain")
    if err == nil {
        t.Fatalf("expected error when minio not initialized")
    }
}
