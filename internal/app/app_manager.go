package app

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/communicator"
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/dedup"
	"NewsFinder/internal/pb/news"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type NewsFinder struct {
	config       Config
	logger       *zap.SugaredLogger
	communicator communicator.Communicator
	dedup        dedup.ManagerDedup
	dm           datamanager.DataManager
	analyzer     analyzer.Analyzer
}

func NewNewsFinder(config Config, logger *zap.SugaredLogger, receiver communicator.Communicator, dm datamanager.DataManager, dedup dedup.ManagerDedup, analyzer analyzer.Analyzer) *NewsFinder {
	return &NewsFinder{
		config:       config,
		logger:       logger,
		communicator: receiver,
		dm:           dm,
		dedup:        dedup,
		analyzer:     analyzer,
	}
}

func (nf *NewsFinder) StartApp() {

}

// StartDataChanWorker TODO: refactor + add locking and parallel workers
func (nf *NewsFinder) StartDataChanWorker() {
	dataChan := nf.communicator.GetConsumeChan()

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

		newsParams, err := nf.convertResultsToNewsParams(message, hardDedupRes, softDedupExists, analysisRes)
		if err != nil {
			nf.logger.Errorw("Error converting message to newsParams, skipping", "message", message)
			continue
		}

		if nf.config.SaveToDB {
			id, err := nf.dm.InsertNews(newsParams)

			if err != nil {
				nf.logger.Errorw("Error inserting news, skipping", "message", message)
				//continue // TODO: решить, будет ли неудачное добавление в бд обрывать пайплайн
			} else {
				nf.logger.Debugw("Inserted new news", "id", id.String())
			}
		}

		if nf.config.ProduceMessages {
			go nf.produceFinalMessage(newsParams)
		}
	}
}

func (nf *NewsFinder) produceFinalMessage(newsParams *datamanager.NewsParams) {
	source, err := nf.dm.GetSourceByID(newsParams.SourceID)
	if err != nil {
		source = datamanager.UnknownSource
		nf.logger.Errorw("Error getting source by id, defaulting to unknown", "id", source.ID.String())
	}

	msg, err := nf.constructProduceMsg(source, newsParams)
	if err != nil {
		nf.logger.Errorw("Error constructing news message, skipping", "message", msg, "err", err)
		return
	}

	nf.communicator.WriteToProduceChan(msg)
}

func (nf *NewsFinder) constructProduceMsg(source *datamanager.Source, newsParams *datamanager.NewsParams) (*communicator.ProduceMessage, error) {
	analysis := &structpb.Struct{}
	if err := protojson.Unmarshal(newsParams.Analysis, analysis); err != nil {
		nf.logger.Errorw("Error unmarshaling analysis", "err", err)
		return nil, err
	}

	pbSource := nf.convertSqlcSourceToPbSource(source)

	newsAnalyzed := news.NewsAnalyzed{
		Source:      pbSource,
		Title:       newsParams.Title,
		Content:     newsParams.Content,
		PublishedAt: newsParams.PublishedAt.Unix(),
		IngestedAt:  newsParams.IngestedAt.Unix(),
		PreparedAt:  time.Now().Unix(),
		Analysis:    analysis,
	}

	return &communicator.ProduceMessage{NewsAnalyzed: &newsAnalyzed}, nil
}
