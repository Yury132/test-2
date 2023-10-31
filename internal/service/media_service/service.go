package media_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/Yury132/Golang-Task-2/internal/models"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nfnt/resize"
	"github.com/rs/zerolog"
)

type MediaService interface {
	CreateThumbnail(info *models.InfoForThumbnail) error
	GetTaskForProcessing() (*models.InfoForThumbnail, error)
}

type ObjectStorage interface {
	Save(data []byte, name string) error
}

type mediaService struct {
	log           zerolog.Logger
	objectStorage ObjectStorage
	jsConsumer    jetstream.Consumer
}

func (m *mediaService) GetTaskForProcessing() (*models.InfoForThumbnail, error) {
	msg, err := m.jsConsumer.Next()
	if err != nil {
		return nil, err
	}

	if err = msg.Ack(); err != nil {
		return nil, err
	}

	if msg.Data() == nil {
		return nil, nil
	}

	var info = new(models.InfoForThumbnail)
	if err = json.Unmarshal(msg.Data(), info); err != nil {
		return nil, err
	}

	return info, nil
}

// Через resize
func (m *mediaService) CreateThumbnail(info *models.InfoForThumbnail) error {

	// Открываем ранее сохраненную картинку
	file, err := os.Open(info.Path)
	if err != nil {
		m.log.Error().Err(err).Msg("failed to open file...")
		return err
	}
	// Получаем image.Image
	imageData, _, err := image.Decode(file)
	if err != nil {
		m.log.Error().Err(err).Msg("failed to decode...")
		return err
	}
	// Закрываем файл
	err = file.Close()
	if err != nil {
		m.log.Error().Err(err).Msg("failed to close file...")
		return err
	}
	// Создаем миниатюру
	newImage := resize.Thumbnail(100, 100, imageData, resize.Lanczos3)

	// Преобразуем в байты
	buf := new(bytes.Buffer)
	err = png.Encode(buf, newImage)
	if err != nil {
		m.log.Error().Err(err).Msg("failed to encode...")
		return err
	}
	imgBytes := buf.Bytes()

	// Создаем уникальное имя
	pName := uuid.New().String()

	if err = m.objectStorage.Save(imgBytes, fmt.Sprintf("%s.png", pName)); err != nil {
		m.log.Error().Err(err).Msg("objectStorage.Save err")
		return err
	}

	return nil
}

// Через vips
// func (m *mediaService) CreateThumbnail(info *models.InfoForThumbnail) error {
// 	img, err := vips.NewImageFromFile(info.Path)

// 	if err != nil {
// 		m.log.Error().Err(err).Msg("NewImageFromFile err")
// 		return err
// 	}

// 	err = img.Thumbnail(info.Size, info.Size, vips.InterestingNone)
// 	if err != nil {
// 		m.log.Error().Err(err).Msg("ResizeWithVScale err")
// 		return err
// 	}

// 	pName := uuid.New().String()

// 	imgBytes, _, err := img.ExportNative()
// 	if err = m.objectStorage.Save(imgBytes, fmt.Sprintf("%s.png", pName)); err != nil {
// 		m.log.Error().Err(err).Msg("objectStorage.Save err")
// 		return err
// 	}

// 	return nil
// }

func New(log zerolog.Logger, objectStorage ObjectStorage, jsConsumer jetstream.Consumer) MediaService {
	return &mediaService{
		log:           log,
		objectStorage: objectStorage,
		jsConsumer:    jsConsumer,
	}
}