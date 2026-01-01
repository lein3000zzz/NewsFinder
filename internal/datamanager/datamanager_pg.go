package datamanager

import (
	"NewsFinder/internal/pb/newsevent"
	"NewsFinder/tools/sqlc/nfsqlc"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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

func (dm *PgDataManager) LookupByHash(hash string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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

func (dm *PgDataManager) InsertNews(event *newsevent.NewsEvent) {
	b, err := proto.Marshal(event)
	if err != nil {
		dm.logger.Errorf("failed to marshal event: %v", err)
		return
	}

	id, err := dm.queries.AddNews(context.Background(), b)
	if err != nil {
		dm.logger.Errorf("failed to insert news: %v", err)
		return
	}

	_ = id
}
