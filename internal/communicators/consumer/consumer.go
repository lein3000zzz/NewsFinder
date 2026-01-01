package consumer

import "NewsFinder/internal/pb/newsevent"

type Consumer interface {
	GetDataChan() <-chan *newsevent.NewsEvent
	StartTopicConsumer()
}
