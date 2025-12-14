package delivery

import (
	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/middleware"
	"backend/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type ExportUsecase interface {
	ExportNoteToPDF(ctx context.Context, userID, noteID uint64) ([]byte, string, error)
}

type ExportDelivery struct {
	usecase ExportUsecase
}

func NewExportDelivery(usecase ExportUsecase) *ExportDelivery {
	return &ExportDelivery{
		usecase: usecase,
	}
}

func (d *ExportDelivery) ExportNoteToPDF(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)

	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	pdf, title, err := d.usecase.ExportNoteToPDF(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to export note to pdf")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	filename := sanitizeFilename(title)

	encodedFilename := url.PathEscape(filename + ".pdf")

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="note.pdf"; filename*=UTF-8''%s`, encodedFilename))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))

	if _, err := w.Write(pdf); err != nil {
		log.Error().Err(err).Msg("failed to write pdf response")
	}
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	result := replacer.Replace(name)

	if len(result) > 100 {
		result = result[:100]
	}

	if result == "" {
		result = "note"
	}

	return result
}
