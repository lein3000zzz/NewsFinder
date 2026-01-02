package datamanager

import (
	"NewsFinder/tools/sqlc/nfsqlc"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type PgDataManager struct {
	logger  *zap.SugaredLogger
	queries *nfsqlc.Queries
}

func NewPgDataManager(logger *zap.SugaredLogger, queries *nfsqlc.Queries) *PgDataManager {
	return &PgDataManager{
		logger:  logger,
		queries: queries,
	}
}

func (dm *PgDataManager) LookupNewsByHash(hash string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := dm.queries.LookupByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			dm.logger.Warnw("no content hash found", "hash", hash)
			return false, nil
		}
		dm.logger.Errorw("error checking content hash", "hash", hash, "err", err)
		return false, err
	}

	return true, nil
}

func (dm *PgDataManager) LookupNewsByEmbedding(vector []float32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	exists, err := dm.queries.LookupEmbedding(ctx, pgvector.NewVector(vector))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			dm.logger.Warnw("no similar content vector embeddings found", "vector", vector)
			return false, nil
		}

		dm.logger.Errorw("error checking content vector embeddings", "vector", vector, "err", err)
		return false, err
	}

	dm.logger.Debugw("content vector embeddings", "vector", vector)

	return exists, nil
}

func (dm *PgDataManager) InsertNews(news *NewsParams) (*uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	newID, err := dm.queries.AddNews(ctx, *news)
	if err != nil {
		dm.logger.Errorw("error adding news", "err", err)
		return nil, err
	}

	return &newID, nil
}

func (dm *PgDataManager) GetSourceByID(id uuid.UUID) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	source, err := dm.queries.GetSourceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			dm.logger.Warnw("no source found", "id", id)
			return nil, ErrNotFound
		}

		dm.logger.Errorw("error getting source by id", "id", id, "err", err)
		return nil, err
	}

	dm.logger.Debugw("source by id", "id", id)

	return &source, nil
}
