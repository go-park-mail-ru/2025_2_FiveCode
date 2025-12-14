package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"backend/pkg/logger"
	"bytes"
	"context"
	"fmt"
	"html"
	"html/template"
	"sort"
)

type PDFGenerator interface {
	ConvertHTMLToPDF(ctx context.Context, html []byte, css []byte) ([]byte, error)
}

type NotesRepository interface {
	GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error)
	GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error)
}

type ExportUsecase struct {
	pdfGenerator PDFGenerator
	notesRepo    NotesRepository
	htmlTemplate *template.Template
	cssStyles    []byte
}

func NewExportUsecase(pdfGenerator PDFGenerator, notesRepo NotesRepository, htmlTemplate *template.Template, cssStyles []byte) *ExportUsecase {
	return &ExportUsecase{
		pdfGenerator: pdfGenerator,
		notesRepo:    notesRepo,
		htmlTemplate: htmlTemplate,
		cssStyles:    cssStyles,
	}
}

type ExportData struct {
	Title   string
	IconURL string
	Blocks  []BlockHTML
}

type BlockHTML struct {
	Type     string
	Content  template.HTML
	Language string
}

func (u *ExportUsecase) ExportNoteToPDF(ctx context.Context, userID, noteID uint64) ([]byte, string, error) {
	log := logger.FromContext(ctx)

	note, err := u.notesRepo.GetNoteById(ctx, userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note")
		return nil, "", fmt.Errorf("failed to get note: %w", err)
	}

	blocks, err := u.notesRepo.GetBlocks(ctx, userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get blocks")
		return nil, "", fmt.Errorf("failed to get blocks: %w", err)
	}

	blocksHTML := make([]BlockHTML, 0, len(blocks))
	for _, block := range blocks {
		blockHTML, err := u.convertBlockToHTML(block)
		if err != nil {
			log.Warn().Err(err).Uint64("block_id", block.ID).Msg("failed to convert block, skipping")
			continue
		}
		blocksHTML = append(blocksHTML, blockHTML)
	}

	exportData := ExportData{
		Title:  note.Title,
		Blocks: blocksHTML,
	}
	if note.Icon != nil {
		exportData.IconURL = note.Icon.URL
	}

	var htmlBuf bytes.Buffer
	if err := u.htmlTemplate.Execute(&htmlBuf, exportData); err != nil {
		log.Error().Err(err).Msg("failed to execute html template")
		return nil, "", fmt.Errorf("failed to render html: %w", err)
	}

	pdf, err := u.pdfGenerator.ConvertHTMLToPDF(ctx, htmlBuf.Bytes(), u.cssStyles)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert html to pdf")
		return nil, "", fmt.Errorf("failed to generate pdf: %w", err)
	}

	log.Info().Uint64("note_id", noteID).Int("pdf_size", len(pdf)).Msg("note exported to pdf")

	return pdf, note.Title, nil
}

func (u *ExportUsecase) convertBlockToHTML(block models.Block) (BlockHTML, error) {
	switch block.Type {
	case models.BlockTypeText:
		content, ok := block.Content.(models.TextContent)
		if !ok {
			if m, ok := block.Content.(map[string]interface{}); ok {
				content = parseTextContentFromMap(m)
			} else {
				return BlockHTML{}, fmt.Errorf("invalid text content type")
			}
		}
		htmlContent := u.renderFormattedText(content)
		return BlockHTML{Type: "text", Content: template.HTML(htmlContent)}, nil

	case models.BlockTypeCode:
		content, ok := block.Content.(models.CodeContent)
		if !ok {
			if m, ok := block.Content.(map[string]interface{}); ok {
				content = parseCodeContentFromMap(m)
			} else {
				return BlockHTML{}, fmt.Errorf("invalid code content type")
			}
		}
		escaped := html.EscapeString(content.Code)
		return BlockHTML{Type: "code", Content: template.HTML(escaped), Language: content.Language}, nil

	case models.BlockTypeAttachment:
		content, ok := block.Content.(models.AttachmentContent)
		if !ok {
			if m, ok := block.Content.(map[string]interface{}); ok {
				content = parseAttachmentContentFromMap(m)
			} else {
				return BlockHTML{}, fmt.Errorf("invalid attachment content type")
			}
		}
		htmlContent := renderAttachment(content)
		return BlockHTML{Type: "attachment", Content: template.HTML(htmlContent)}, nil

	default:
		return BlockHTML{}, fmt.Errorf("unknown block type: %s", block.Type)
	}
}

func (u *ExportUsecase) renderFormattedText(content models.TextContent) string {
	if len(content.Formats) == 0 {
		return "<p>" + html.EscapeString(content.Text) + "</p>"
	}

	formats := make([]models.BlockTextFormat, len(content.Formats))
	copy(formats, content.Formats)
	sort.Slice(formats, func(i, j int) bool {
		return formats[i].StartOffset < formats[j].StartOffset
	})

	text := []rune(content.Text)
	var result bytes.Buffer
	result.WriteString("<p>")

	pos := 0
	for _, f := range formats {
		if f.StartOffset > pos {
			result.WriteString(html.EscapeString(string(text[pos:f.StartOffset])))
		}

		end := f.EndOffset
		if end > len(text) {
			end = len(text)
		}

		chunk := html.EscapeString(string(text[f.StartOffset:end]))
		styled := applyStyles(chunk, f)
		result.WriteString(styled)

		pos = end
	}

	if pos < len(text) {
		result.WriteString(html.EscapeString(string(text[pos:])))
	}

	result.WriteString("</p>")
	return result.String()
}

func applyStyles(text string, f models.BlockTextFormat) string {
	var styles []string

	if f.Font != "" && f.Font != models.FontInter {
		styles = append(styles, fmt.Sprintf("font-family: '%s', sans-serif", f.Font))
	}
	if f.Size != 0 && f.Size != 12 {
		styles = append(styles, fmt.Sprintf("font-size: %dpx", f.Size))
	}

	result := text

	if f.Bold {
		result = "<strong>" + result + "</strong>"
	}
	if f.Italic {
		result = "<em>" + result + "</em>"
	}
	if f.Underline {
		result = "<u>" + result + "</u>"
	}
	if f.Strikethrough {
		result = "<s>" + result + "</s>"
	}
	if f.Link != nil && *f.Link != "" {
		result = fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(*f.Link), result)
	}

	if len(styles) > 0 {
		result = fmt.Sprintf(`<span style="%s">%s</span>`, joinStyles(styles), result)
	}

	return result
}

func joinStyles(styles []string) string {
	var result string
	for i, s := range styles {
		if i > 0 {
			result += "; "
		}
		result += s
	}
	return result
}

func renderAttachment(content models.AttachmentContent) string {
	if isImage(content.MimeType) {
		caption := ""
		if content.Caption != nil {
			caption = fmt.Sprintf(`<figcaption>%s</figcaption>`, html.EscapeString(*content.Caption))
		}
		return fmt.Sprintf(`<figure><img src="%s" alt="attachment">%s</figure>`, html.EscapeString(content.URL), caption)
	}

	caption := "Download attachment"
	if content.Caption != nil {
		caption = *content.Caption
	}
	return fmt.Sprintf(`<p class="file-link"><a href="%s">📎 %s</a></p>`, html.EscapeString(content.URL), html.EscapeString(caption))
}

func isImage(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml":
		return true
	}
	return false
}

func parseTextContentFromMap(m map[string]interface{}) models.TextContent {
	var content models.TextContent
	if text, ok := m["text"].(string); ok {
		content.Text = text
	}
	if formats, ok := m["formats"].([]interface{}); ok {
		for _, f := range formats {
			if fm, ok := f.(map[string]interface{}); ok {
				content.Formats = append(content.Formats, parseFormatFromMap(fm))
			}
		}
	}
	return content
}

func parseFormatFromMap(m map[string]interface{}) models.BlockTextFormat {
	var f models.BlockTextFormat
	if v, ok := m["start_offset"].(float64); ok {
		f.StartOffset = int(v)
	}
	if v, ok := m["end_offset"].(float64); ok {
		f.EndOffset = int(v)
	}
	if v, ok := m["bold"].(bool); ok {
		f.Bold = v
	}
	if v, ok := m["italic"].(bool); ok {
		f.Italic = v
	}
	if v, ok := m["underline"].(bool); ok {
		f.Underline = v
	}
	if v, ok := m["strikethrough"].(bool); ok {
		f.Strikethrough = v
	}
	if v, ok := m["link"].(string); ok && v != "" {
		f.Link = &v
	}
	if v, ok := m["font"].(string); ok {
		f.Font = models.TextFont(v)
	}
	if v, ok := m["size"].(float64); ok {
		f.Size = int(v)
	}
	return f
}

func parseCodeContentFromMap(m map[string]interface{}) models.CodeContent {
	var content models.CodeContent
	if v, ok := m["code"].(string); ok {
		content.Code = v
	}
	if v, ok := m["language"].(string); ok {
		content.Language = v
	}
	return content
}

func parseAttachmentContentFromMap(m map[string]interface{}) models.AttachmentContent {
	var content models.AttachmentContent
	if v, ok := m["url"].(string); ok {
		content.URL = v
	}
	if v, ok := m["caption"].(string); ok {
		content.Caption = &v
	}
	if v, ok := m["mime_type"].(string); ok {
		content.MimeType = v
	}
	return content
}
