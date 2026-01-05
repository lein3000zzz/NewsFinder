package app

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/communicator"
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/dedup"
	"NewsFinder/internal/pb/news"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type sourceWorker struct {
	sync.Mutex
}

type NewsFinder struct {
	config       Config
	logger       *zap.SugaredLogger
	communicator communicator.Communicator
	dedup        dedup.ManagerDedup
	dm           datamanager.DataManager
	analyzer     analyzer.Analyzer
}

func NewNewsFinder(
	config Config,
	logger *zap.SugaredLogger,
	communicator communicator.Communicator,
	dm datamanager.DataManager,
	dedup dedup.ManagerDedup,
	analyzer analyzer.Analyzer,
) *NewsFinder {
	return &NewsFinder{
		config:       config,
		logger:       logger,
		communicator: communicator,
		dm:           dm,
		dedup:        dedup,
		analyzer:     analyzer,
	}
}

func (nf *NewsFinder) StartApp() {
	go nf.communicator.StartTopicConsumer()
	go nf.communicator.StartTopicProducer()
	go nf.startDataChanWorker()
}

func (nf *NewsFinder) startDataChanWorker() {
	nf.logger.Infow("starting data channel worker")

	dataChan := nf.communicator.GetConsumeChan()

	workers := make(map[string]*sourceWorker)
	var workersMu sync.RWMutex

	var wg sync.WaitGroup
	for i := 0; i < nf.config.WorkerCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for message := range dataChan {
				sourceID := message.Event.SourceId

				sw := nf.processSource(sourceID, workers, &workersMu)

				nf.processMessageBySourceWorker(sw, message)
			}
		}()
	}

	wg.Wait()

	nf.logger.Info("All workers have been stopped")
}

func (nf *NewsFinder) processSource(sourceID string, workers map[string]*sourceWorker, workersMu *sync.RWMutex) *sourceWorker {
	sw, exists := nf.getSourceWorker(sourceID, workers, workersMu)
	if !exists {
		sw = nf.processNewSource(sourceID, workers, workersMu)
	}
	return sw
}

func (nf *NewsFinder) processNewSource(sourceID string, workers map[string]*sourceWorker, workersMu *sync.RWMutex) *sourceWorker {
	workersMu.Lock()
	defer workersMu.Unlock()

	sw, exists := workers[sourceID]
	if !exists {
		sw = &sourceWorker{}
		workers[sourceID] = sw
	}

	return sw
}

func (nf *NewsFinder) getSourceWorker(sourceID string, workers map[string]*sourceWorker, workersMu *sync.RWMutex) (*sourceWorker, bool) {
	workersMu.RLock()
	defer workersMu.RUnlock()

	sw, exists := workers[sourceID]
	return sw, exists
}

func (nf *NewsFinder) processMessageBySourceWorker(sw *sourceWorker, message *communicator.ConsumeMessage) {
	sw.Lock()
	defer sw.Unlock()

	nf.processMessage(message)
}

func (nf *NewsFinder) processMessage(message *communicator.ConsumeMessage) {
	nf.logger.Infow("received datachan message", "message", message)

	hardDedupRes, err := nf.dedup.CheckExistsHard(message.Event)
	if err != nil {
		nf.logger.Errorw("Error checking existence hard, skipping", "err", err)
		return
	}

	if hardDedupRes.Exists {
		nf.logger.Infow("Found hard duplicate, skipping", "message", message)
		return
	}

	softDedupRes, err := nf.dedup.CheckExistsSoft(hardDedupRes)
	if err != nil {
		nf.logger.Errorw("Error checking existence soft, skipping", "err", err)
		return
	}

	if softDedupRes.Exists {
		nf.logger.Infow("Found soft duplicate, but not skipping future processing", "message", message)
	}

	analysisRes, err := nf.analyzer.Analyze(hardDedupRes.Normalized)
	if err != nil {
		nf.logger.Errorw("Error analyzing hard duplicate, skipping", "err", err)
		return
	}

	newsParams, err := nf.convertResultsToNewsParams(message, hardDedupRes, softDedupRes, analysisRes)
	if err != nil {
		nf.logger.Errorw("Error converting message to newsParams, skipping", "message", message)
		return
	}

	id, err := nf.dm.InsertNews(newsParams)
	if err != nil {
		nf.logger.Errorw("Error inserting news, skipping", "message", message)
		//return // TODO: решить, будет ли неудачное добавление в бд обрывать пайплайн
	} else {
		nf.logger.Debugw("Inserted new news", "id", id.String())
	}

	if nf.config.ProduceMessages {
		go nf.produceFinalMessage(newsParams)
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
