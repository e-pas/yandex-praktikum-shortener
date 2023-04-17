package saver

import (
	"context"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
)

type mediaSaver interface {
	Save(ctx context.Context, data *config.ShortURL) error
	SaveBatch(ctx context.Context, data map[string]*config.ShortURL) error
	Load(ctx context.Context, data map[string]*config.ShortURL) error
	Ping(ctx context.Context) error
}

type Saver struct {
	ms mediaSaver
}

func New(c *config.Config) *Saver {
	ms := mediaSaver(nil)
	if c.PgConnString != "" {
		ms = newPgSaver(c.PgConnString)
	} else {
		if c.FileStorage != "" {
			ms = newDiskSaver(c.FileStorage)
		}
	}
	return &Saver{
		ms: ms,
	}
}

func (s *Saver) Save(ctx context.Context, data *config.ShortURL) error {
	if s.ms != nil {
		return s.ms.Save(ctx, data)
	}
	return nil
}

func (s *Saver) SaveBatch(ctx context.Context, data map[string]*config.ShortURL) error {
	if s.ms != nil {
		return s.ms.SaveBatch(ctx, data)
	}
	return nil
}

func (s *Saver) Load(ctx context.Context, data map[string]*config.ShortURL) error {
	if s.ms != nil {
		return s.ms.Load(ctx, data)
	}
	return nil
}

func (s *Saver) Ping(ctx context.Context) error {
	if s.ms != nil {
		return s.ms.Ping(ctx)
	}
	return config.ErrNoAttachedStorage
}
