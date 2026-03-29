package repository

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/admin/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
)

type repository struct {
	queries *client.Queries
}

func New(q *client.Queries) service.Repository {
	return &repository{queries: q}
}

func (r *repository) CreateIfNotExists(ctx context.Context, id int32) error {
	return logger.WrapError(
		r.queries.CreateAdminIfNotExists(ctx, id),
	)
}
