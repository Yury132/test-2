package postgres

import (
	"context"
	"time"

	"github.com/Yury132/Golang-Task-2/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Storage interface {
	SaveFileMeta(ctx context.Context, metaInfo *models.ImageMeta) error
	SaveFileMiniMeta(ctx context.Context, metaInfo *models.ImageMeta) error
}

type storage struct {
	conn *pgxpool.Pool
}

// Загрузка данных в БД об изначальных изображениях
func (s *storage) SaveFileMeta(ctx context.Context, metaInfo *models.ImageMeta) error {
	query := "INSERT INTO public.uploads_info (name, type, width, height) VALUES ($1, $2, $3, $4)"

	ctxDb, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.conn.Exec(ctxDb, query, metaInfo.Name, metaInfo.Type, metaInfo.Width, metaInfo.Height)
	if err != nil {
		return errors.Wrap(err, "failed to write file meta to db")
	}

	return nil
}

// Загрузка данных в БД о миниатюрах
func (s *storage) SaveFileMiniMeta(ctx context.Context, metaInfo *models.ImageMeta) error {
	query := "INSERT INTO public.mini_info (name, type, width, height) VALUES ($1, $2, $3, $4)"

	ctxDb, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.conn.Exec(ctxDb, query, metaInfo.Name, metaInfo.Type, metaInfo.Width, metaInfo.Height)
	if err != nil {
		return errors.Wrap(err, "failed to write fileMini meta to db")
	}

	return nil
}

func New(conn *pgxpool.Pool) Storage {
	return &storage{
		conn: conn,
	}
}
