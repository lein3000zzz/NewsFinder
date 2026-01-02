package datamanager

import (
	"NewsFinder/tools/sqlc/nfsqlc"
	"time"

	"github.com/google/uuid"
)

const (
	dbTimeout = 3 * time.Second
)

type News = nfsqlc.News
type NewsParams = nfsqlc.AddNewsParams
type Source = nfsqlc.Source

var (
	UnknownSource = &Source{
		Name:        "unknown",
		Credibility: 0,
	}
)

type DataManager interface {
	LookupNewsByHash(hash string) (bool, error)
	LookupNewsByEmbedding(vector []float32) (bool, error)
	InsertNews(news *NewsParams) (*uuid.UUID, error)
	GetSourceByID(id uuid.UUID) (*Source, error)
}
