package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"backend/models"

	"github.com/stretchr/testify/assert"
)

type fu_fakeRepo struct{
    uploadedURL string
    saveErr error
    files map[uint64]*models.File
    deleteMinIOErr error
}

func (f *fu_fakeRepo) UploadFileToMinIO(ctx context.Context, filename string, fileData []byte, contentType string) (string, error) {
    if f.uploadedURL == "" {
        return "http://example.com/"+filename, nil
    }
    return f.uploadedURL, nil
}
func (f *fu_fakeRepo) SaveFile(ctx context.Context, url, mimeType string, sizeBytes int64, width, height *int) (*models.File, error) {
    if f.saveErr != nil {
        return nil, f.saveErr
    }
    id := uint64(len(f.files)+1)
    file := &models.File{ID:id, URL:url, MimeType:mimeType, SizeBytes:sizeBytes}
    if width != nil { w := *width; file.Width = &w }
    if height != nil { h := *height; file.Height = &h }
    if f.files == nil { f.files = map[uint64]*models.File{} }
    f.files[id] = file
    return file, nil
}
func (f *fu_fakeRepo) GetFileByID(ctx context.Context, fileID uint64) (*models.File, error) {
    if file, ok := f.files[fileID]; ok { return file, nil }
    return nil, errors.New("not found")
}
func (f *fu_fakeRepo) DeleteFile(ctx context.Context, fileID uint64) error {
    if _, ok := f.files[fileID]; ok { delete(f.files, fileID); return nil }
    return errors.New("not found")
}
func (f *fu_fakeRepo) DeleteFileFromMinIO(ctx context.Context, url string) error { return f.deleteMinIOErr }

func TestFileUsecase_UploadAndGet_DeleteFlow(t *testing.T) {
    repo := &fu_fakeRepo{}
    u := NewFileUsecase(repo)

    // non-image upload
    r := bytes.NewReader([]byte("hello"))
    fileModel, err := u.UploadFile(context.Background(), io.NopCloser(r), "file.txt", "text/plain", 5)
    assert.NoError(t, err)
    assert.NotNil(t, fileModel)

    // get file
    got, err := u.GetFile(context.Background(), fileModel.ID)
    assert.NoError(t, err)
    assert.Equal(t, fileModel.URL, got.URL)

    // delete file
    err = u.DeleteFile(context.Background(), fileModel.ID)
    assert.NoError(t, err)
}

func TestFileUsecase_Upload_ImageDecodingAndSaveFailCleanup(t *testing.T) {
    // create a fake repo that will fail SaveFile to test cleanup path
    repo := &fu_fakeRepo{saveErr: errors.New("save failed")}
    u := NewFileUsecase(repo)

    // small valid PNG header to make image.DecodeConfig succeed
    pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
    r := bytes.NewReader(pngBytes)

    _, err := u.UploadFile(context.Background(), io.NopCloser(r), "image.png", "image/png", int64(len(pngBytes)))
    assert.Error(t, err)
}

func TestFileUsecase_GetFile_Error(t *testing.T) {
    repo := &fu_fakeRepo{}
    u := NewFileUsecase(repo)

    _, err := u.GetFile(context.Background(), 999)
    assert.Error(t, err)
}

func TestFileUsecase_DeleteFile_DeleteFromMinIOError(t *testing.T) {
    // simulate delete-from-minio error but DeleteFile still succeeds
    repo := &fu_fakeRepo{}
    u := NewFileUsecase(repo)

    // create file via SaveFile
    saved, _ := repo.SaveFile(context.Background(), "http://example.com/x", "text/plain", 1, nil, nil)

    // configure fake to return error on DeleteFileFromMinIO
    repo.deleteMinIOErr = errors.New("minio fail")

    err := u.DeleteFile(context.Background(), saved.ID)
    assert.NoError(t, err)
}
