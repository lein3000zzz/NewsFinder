package app

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/communicator"
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/dedup"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type NewsFinder struct {
	logger   *zap.SugaredLogger
	consumer communicator.Communicator
	dedup    dedup.ManagerDedup
	dm       datamanager.DataManager
	analyzer analyzer.Analyzer
}

func NewNewsFinder(logger *zap.SugaredLogger, receiver communicator.Communicator, dm datamanager.DataManager, dedup dedup.ManagerDedup, analyzer analyzer.Analyzer) *NewsFinder {
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
	dataChan := nf.consumer.GetConsumeChan()

	for message := range dataChan {
		hardDedupRes, err := nf.dedup.CheckExistsHard(message.Event)
		if err != nil {
			nf.logger.Errorw("Error checking existence hard, skipping", "err", err)
			continue
		}

		if hardDedupRes.Exists {
			nf.logger.Debugw("Found hard duplicate, skipping...", "message", message)
			continue
		}

		softDedupExists, err := nf.dedup.CheckExistsSoft(hardDedupRes)
		if err != nil {
			nf.logger.Errorw("Error checking existence soft, skipping", "err", err)
			continue
		}

		analysisRes, err := nf.analyzer.Analyze(hardDedupRes.Normalized)
		if err != nil {
			nf.logger.Errorw("Error analyzing hard duplicate, skipping", "err", err)
			continue
		}

		news, err := nf.convertDataToAddNews(message, hardDedupRes, softDedupExists, analysisRes)
		if err != nil {
			nf.logger.Errorw("Error converting message to news, skipping", "message", message)
			continue
		}

		id, err := nf.dm.InsertNews(news)
		if err != nil {
			nf.logger.Errorw("Error inserting news, skipping", "message", message)
			continue
		}

		nf.logger.Debugw("Inserted new news", "id", id.String())
	}
}

func (nf *NewsFinder) convertDataToAddNews(
	message *communicator.ConsumeMessage,
	hardDedupRes *dedup.HardDedupResult,
	softDedupRes *dedup.SoftDedupRes,
	analysisRes *analyzer.AnalysisRes,
) (*datamanager.NewsParams, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		nf.logger.Errorw("Error generating new v7", "err", err)
		return nil, err
	}

	sourceID, err := uuid.Parse(message.Event.SourceId)
	if err != nil {
		nf.logger.Errorw("Error parsing source id", "err", err)
		return nil, err
	}

	analysis, err := json.Marshal(analysisRes)
	if err != nil {
		nf.logger.Errorw("Error marshalling analysis", "err", err)
		return nil, err
	}

	news := &datamanager.NewsParams{
		ID:               v7,
		SourceID:         sourceID,
		Title:            message.Event.Title,
		Content:          message.Event.Content,
		PublishedAt:      time.Unix(message.Event.PublishedAt, 0),
		IngestedAt:       message.IngestedAt,
		ContentHash:      hardDedupRes.Hash,
		ContentEmbedding: pgvector.NewVector(softDedupRes.Vector),
		Analysis:         analysis,
	}

	nf.logger.Debugw("Converted news", "news", news)

	return news, nil
}
