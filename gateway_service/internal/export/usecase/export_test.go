package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"context"
	"errors"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPDFGenerator struct {
	pdf []byte
	err error
}

func (m *mockPDFGenerator) ConvertHTMLToPDF(ctx context.Context, html []byte, css []byte) ([]byte, error) {
	return m.pdf, m.err
}

type mockNotesRepository struct {
	note         *models.Note
	blocks       []models.Block
	getNoteErr   error
	getBlocksErr error
}

func (m *mockNotesRepository) GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error) {
	return m.note, m.getNoteErr
}

func (m *mockNotesRepository) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error) {
	return m.blocks, m.getBlocksErr
}

func TestNewExportUsecase(t *testing.T) {
	pdfGen := &mockPDFGenerator{}
	notesRepo := &mockNotesRepository{}
	tmpl, _ := template.New("test").Parse("<html>{{.Title}}</html>")
	css := []byte("body { color: black; }")

	usecase := NewExportUsecase(pdfGen, notesRepo, tmpl, css)
	assert.NotNil(t, usecase)
	assert.Equal(t, pdfGen, usecase.pdfGenerator)
	assert.Equal(t, notesRepo, usecase.notesRepo)
}

func TestExportUsecase_ExportNoteToPDF(t *testing.T) {
	tmpl, _ := template.New("note").Parse(`
		<html>
			<head><style>{{.}}</style></head>
			<body>
				<h1>{{.Title}}</h1>
				{{range .Blocks}}
					<div class="{{.Type}}">{{.Content}}</div>
				{{end}}
			</body>
		</html>
	`)
	css := []byte("body { margin: 0; }")

	t.Run("Success", func(t *testing.T) {
		pdfGen := &mockPDFGenerator{
			pdf: []byte("fake pdf"),
			err: nil,
		}
		notesRepo := &mockNotesRepository{
			note: &models.Note{
				ID:    1,
				Title: "Test Note",
			},
			blocks: []models.Block{
				{
					ID:   1,
					Type: models.BlockTypeText,
					Content: models.TextContent{
						Text: "Hello world",
					},
				},
			},
		}

		usecase := NewExportUsecase(pdfGen, notesRepo, tmpl, css)
		ctx := context.Background()

		pdf, title, err := usecase.ExportNoteToPDF(ctx, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, []byte("fake pdf"), pdf)
		assert.Equal(t, "Test Note", title)
	})

	t.Run("GetNoteError", func(t *testing.T) {
		pdfGen := &mockPDFGenerator{}
		notesRepo := &mockNotesRepository{
			getNoteErr: errors.New("note not found"),
		}

		usecase := NewExportUsecase(pdfGen, notesRepo, tmpl, css)
		ctx := context.Background()

		_, _, err := usecase.ExportNoteToPDF(ctx, 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get note")
	})

	t.Run("GetBlocksError", func(t *testing.T) {
		pdfGen := &mockPDFGenerator{}
		notesRepo := &mockNotesRepository{
			note: &models.Note{
				ID:    1,
				Title: "Test Note",
			},
			getBlocksErr: errors.New("blocks error"),
		}

		usecase := NewExportUsecase(pdfGen, notesRepo, tmpl, css)
		ctx := context.Background()

		_, _, err := usecase.ExportNoteToPDF(ctx, 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get blocks")
	})

	t.Run("PDFGenerationError", func(t *testing.T) {
		pdfGen := &mockPDFGenerator{
			err: errors.New("pdf generation error"),
		}
		notesRepo := &mockNotesRepository{
			note: &models.Note{
				ID:    1,
				Title: "Test Note",
			},
			blocks: []models.Block{},
		}

		usecase := NewExportUsecase(pdfGen, notesRepo, tmpl, css)
		ctx := context.Background()

		_, _, err := usecase.ExportNoteToPDF(ctx, 1, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate pdf")
	})
}

func TestIsImage(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"image/svg+xml", true},
		{"text/plain", false},
		{"application/pdf", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := isImage(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinStyles(t *testing.T) {
	t.Run("MultipleStyles", func(t *testing.T) {
		styles := []string{"color: red", "font-size: 12px", "margin: 0"}
		result := joinStyles(styles)
		assert.Equal(t, "color: red; font-size: 12px; margin: 0", result)
	})

	t.Run("SingleStyle", func(t *testing.T) {
		styles := []string{"color: red"}
		result := joinStyles(styles)
		assert.Equal(t, "color: red", result)
	})

	t.Run("Empty", func(t *testing.T) {
		styles := []string{}
		result := joinStyles(styles)
		assert.Equal(t, "", result)
	})
}
