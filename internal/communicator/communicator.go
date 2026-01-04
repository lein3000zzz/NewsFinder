package communicator

import (
	"NewsFinder/internal/pb/news"
	"time"
)

type ConsumeMessage struct {
	Event      *news.NewsEvent
	IngestedAt time.Time
}

type ProduceMessage struct {
	NewsAnalyzed *news.NewsAnalyzed
}

type Communicator interface {
	GetConsumeChan() <-chan *ConsumeMessage
	StartTopicConsumer()
	StartTopicProducer()
	WriteToProduceChan(msg *ProduceMessage)
}
