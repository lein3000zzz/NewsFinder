package app

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/communicator"
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/dedup"
	"NewsFinder/internal/pb/news"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

func (nf *NewsFinder) convertResultsToNewsParams(
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

	newsParams := &datamanager.NewsParams{
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

	nf.logger.Debugw("Converted newsParams", "newsParams", newsParams)

	return newsParams, nil
}

func (nf *NewsFinder) convertSqlcSourceToPbSource(source *datamanager.Source) *news.Source {
	return &news.Source{
		Name:        source.Name,
		Credibility: source.Credibility,
	}
}
