package repo

import (
	"context"

	"feedback-bot/internal/entity"
	"feedback-bot/internal/repo/pgdb"
	"feedback-bot/pkg/postgres"
)

type MessageUpdate interface {
	Create(ctx context.Context, entity entity.MessageUpdate) (int, error)
	GetById(ctx context.Context, id int) (entity.MessageUpdate, error)
	GetList(ctx context.Context) ([]entity.MessageUpdate, error)
}

type UserThread interface {
	Add(ctx context.Context, entity entity.UserThread) error
	// Edit(ctx context.Context, postId int64, fwrdId int64) error
	GetList(ctx context.Context) (map[int64]int64, error)
}

type (
	Repositories struct {
		MessageUpdate
		UserThread
		// ...
	}
)

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		MessageUpdate: pgdb.NewRepoMessageUpdate(pg),
		UserThread:    pgdb.NewRepoUserThread(pg),
		// ...
	}
}
