package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/Yury132/Golang-Task-2/internal/models"
	"github.com/rs/zerolog"
)

type Service interface {
	UploadPhoto(ctx context.Context, data []byte, metaInfo *models.ImageMeta, thumbSize int) error
	// Получаем информацию о картинках
	GetData(ctx context.Context) ([]models.AllImages, error)
}

type Handler struct {
	log     zerolog.Logger
	service Service
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := "{\"health\": \"ok\"}"

	response, err := json.Marshal(data)
	if err != nil {
		h.log.Error().Err(err).Msg("filed to marshal response data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	scaleStr := queryParams.Get("size")
	size, err := strconv.Atoi(scaleStr)
	if err != nil {
		h.log.Error().Err(err).Msg("invalid query param - size")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		h.log.Error().Err(err).Msg("failed to upload file")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		if err = file.Close(); err != nil {
			h.log.Error().Err(err).Send()
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil || data == nil {
		h.log.Error().Err(err).Msg("failed to read the file")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metaInfo, err := models.CollectImageMeta(data, handler.Filename)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to collect meta info")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = h.service.UploadPhoto(r.Context(), data, metaInfo, size); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Получаем информацию о картинках
func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	images, err := h.service.GetData(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error().Err(err).Msg("failed to get images")
		return
	}
	// Кодируем
	data, err := json.Marshal(images)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error().Err(err).Msg("failed to marshal images")
		return
	}
	w.Write(data)
}

func New(log zerolog.Logger, service Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}
