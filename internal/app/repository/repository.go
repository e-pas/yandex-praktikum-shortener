package repository

import (
	"context"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/model"
)

type mediaRepository interface {
	Save(ctx context.Context, data model.ShortURL) error
	SaveBatch(ctx context.Context, data []*model.ShortURL) error
	UpdateBatch(ctx context.Context, data []*model.ShortURL) error
	Load(ctx context.Context, data map[string]*model.ShortURL) error
	Ping(ctx context.Context) error
}

type Repository struct {
	ms mediaRepository
}

func New(c *config.Config) *Repository {
	ms := mediaRepository(nil)
	if c.PgConnString != "" {
		ms = newPgSaver(c.PgConnString)
	} else {
		if c.FileStorage != "" {
			ms = newDiskSaver(c.FileStorage)
		}
	}
	return &Repository{
		ms: ms,
	}
}

func (s *Repository) Save(ctx context.Context, data model.ShortURL) error {
	if s.ms != nil {
		return s.ms.Save(ctx, data)
	}
	return nil
}

func (s *Repository) SaveBatch(ctx context.Context, data []*model.ShortURL) error {
	if s.ms != nil {
		return s.ms.SaveBatch(ctx, data)
	}
	return nil
}

func (s *Repository) UpdateBatch(ctx context.Context, data []*model.ShortURL) error {
	if s.ms != nil {
		return s.ms.UpdateBatch(ctx, data)
	}
	return nil
}

func (s *Repository) Load(ctx context.Context, data map[string]*model.ShortURL) error {
	if s.ms != nil {
		return s.ms.Load(ctx, data)
	}
	return nil
}

func (s *Repository) Ping(ctx context.Context) error {
	if s.ms != nil {
		return s.ms.Ping(ctx)
	}
	return nil
}
