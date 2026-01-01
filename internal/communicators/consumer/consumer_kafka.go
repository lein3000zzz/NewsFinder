package consumer

import (
	"NewsFinder/internal/pb/newsevent"
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type KafkaConsumer struct {
	logger      *zap.SugaredLogger
	kafkaClient *kgo.Client
	dataChan    chan *newsevent.NewsEvent
}

func NewKafkaConsumer(logger *zap.SugaredLogger, seeds []string, consumerGroup, topic string) *KafkaConsumer {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		logger.Fatal(err)
	}

	return &KafkaConsumer{
		logger:      logger,
		kafkaClient: client,
		dataChan:    make(chan *newsevent.NewsEvent),
	}
}

func (kr *KafkaConsumer) GetDataChan() <-chan *newsevent.NewsEvent {
	return kr.dataChan
}

func (kr *KafkaConsumer) StartTopicConsumer() {
	ctx := context.Background()

	for {
		fetches := kr.kafkaClient.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			kr.logger.Errorw("Error polling fetches", "errors", errs)
			continue
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			var newsEvent newsevent.NewsEvent
			err := proto.Unmarshal(record.Value, &newsEvent)
			if err != nil {
				kr.logger.Errorw("Error unmarshalling news event", "record", record)
				continue
			}

			kr.logger.Debugw("parsed news event", &newsEvent)
			kr.dataChan <- &newsEvent
		}
	}
}
