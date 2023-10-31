package main_service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yury132/Golang-Task-2/internal/models"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
)

type Storage interface {
	SaveFileMeta(ctx context.Context, metaInfo *models.ImageMeta) error
}

type ObjectStorage interface {
	Save(data []byte, name string) error
}

type Service interface {
	UploadPhoto(ctx context.Context, data []byte, metaInfo *models.ImageMeta, thumbSize int) error
}

type service struct {
	log           zerolog.Logger
	storage       Storage
	objectStorage ObjectStorage
	js            jetstream.JetStream
}

func (s *service) UploadPhoto(ctx context.Context, data []byte, metaInfo *models.ImageMeta, thumbSize int) error {
	if err := s.objectStorage.Save(data, metaInfo.Name); err != nil {
		s.log.Error().Err(err).Msg("save to object storage err")
		return err
	}

	if err := s.storage.SaveFileMeta(ctx, metaInfo); err != nil {
		s.log.Error().Err(err).Msg("save to db err")
		return err
	}

	msg := models.InfoForThumbnail{
		Path: fmt.Sprintf("uploads/%s", metaInfo.Name),
		Size: thumbSize,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		s.log.Error().Err(err).Msg("js message marshal err")
		return err
	}

	if _, err = s.js.Publish(ctx, "media.picture", b); err != nil {
		s.log.Error().Err(err).Msg("failed to publish message")
		return err
	}

	return nil
}

func New(log zerolog.Logger, storage Storage, objectStorage ObjectStorage, js jetstream.JetStream) Service {
	return &service{
		log:           log,
		storage:       storage,
		objectStorage: objectStorage,
		js:            js,
	}
}
