package app

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/communicators/consumer"
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/dedup"
	"NewsFinder/internal/pb/newsevent"
	"NewsFinder/tools/sqlc/nfsqlc"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type NewsFinder struct {
	logger   *zap.SugaredLogger
	consumer consumer.Consumer
	dedup    dedup.ManagerDedup
	dm       datamanager.DataManager
	analyzer analyzer.Analyzer
}

func NewNewsFinder(logger *zap.SugaredLogger, receiver consumer.Consumer, dm datamanager.DataManager, dedup dedup.ManagerDedup, analyzer analyzer.Analyzer) *NewsFinder {
	return &NewsFinder{
		logger:   logger,
		consumer: receiver,
		dm:       dm,
		dedup:    dedup,
		analyzer: analyzer,
	}
}

func (nf *NewsFinder) StartApp() {

}

func (nf *NewsFinder) StartDataChanWorker() {
	dataChan := nf.consumer.GetDataChan()

	for event := range dataChan {
		hardDedupRes, err := nf.dedup.CheckExistsHard(event)
		if err != nil {
			nf.logger.Errorw("Error checking existence hard, skipping", "err", err)
			continue
		}

		if hardDedupRes.Exists {
			nf.logger.Debugw("Found hard duplicate, skipping...", "event", event)
			continue
		}

		softDedupExists, err := nf.dedup.CheckExistsSoft(hardDedupRes)
		if err != nil {
			nf.logger.Errorw("Error checking existence soft, skipping", "err", err)
			continue
		}

		news, err := nf.convertEventToNews(event, hardDedupRes, softDedupExists)
		if err != nil {
			nf.logger.Errorw("Error converting event to news, skipping", "event", event)
			continue
		}

	}
}

// Конверт инициализирует стандартные значения для полей Analysis ([]byte), ContentEmbedding (pgvector.Vector), CreatedAt (time.Time)
// они будут заполнены далее в пайплайне
func (nf *NewsFinder) convertEventToNews(event *newsevent.NewsEvent, hardDedupRes *dedup.HardDedupResult, softDedupRes *dedup.SoftDedupRes) (*nfsqlc.News, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		nf.logger.Errorw("Error generating new v7", "err", err)
		return nil, err
	}

	sourceID, err := uuid.Parse(event.SourceId)
	if err != nil {
		nf.logger.Errorw("Error parsing source id", "err", err)
		return nil, err
	}

	news := &nfsqlc.News{
		ID:               v7,
		SourceID:         sourceID,
		Title:            event.Title,
		Content:          event.Content,
		PublishedAt:      time.Unix(event.PublishedAt, 0),
		IngestedAt:       time.Now(),
		ContentHash:      hardDedupRes.Hash,
		ContentEmbedding: pgvector.NewVector(softDedupRes.Vector),
	}

	return news, nil
}
