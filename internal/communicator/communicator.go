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
	Data []byte
}

type Communicator interface {
	GetConsumeChan() <-chan *ConsumeMessage
	StartTopicConsumer()
}
