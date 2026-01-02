package communicator

import (
	"NewsFinder/internal/pb/news"
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type KafkaConfig struct {
	Seeds         []string
	ConsumerGroup string
	Topic         string
	User          string
	Password      string
}

type KafkaCommunicator struct {
	logger      *zap.SugaredLogger
	kafkaClient *kgo.Client
	consumeChan chan *ConsumeMessage
	produceChan chan *ProduceMessage
}

func NewKafkaConsumer(logger *zap.SugaredLogger, config *KafkaConfig) *KafkaCommunicator {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.Seeds...),
		kgo.ConsumerGroup(config.ConsumerGroup),
		kgo.ConsumeTopics(config.Topic),
		kgo.SASL(plain.Auth{
			User: config.User,
			Pass: config.Password,
		}.AsMechanism()),
	)
	if err != nil {
		logger.Fatal(err)
	}

	return &KafkaCommunicator{
		logger:      logger,
		kafkaClient: client,
		consumeChan: make(chan *ConsumeMessage),
		produceChan: make(chan *ProduceMessage),
	}
}

func (kr *KafkaCommunicator) GetConsumeChan() <-chan *ConsumeMessage {
	return kr.consumeChan
}

func (kr *KafkaCommunicator) GetProduceChan() <-chan *ProduceMessage {
	return kr.produceChan
}

func (kr *KafkaCommunicator) StartTopicConsumer() {
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

			var newsEvent news.NewsEvent
			err := proto.Unmarshal(record.Value, &newsEvent)
			if err != nil {
				kr.logger.Errorw("Error unmarshalling news event", "record", record)
				continue
			}

			message := &ConsumeMessage{
				Event:      &newsEvent,
				IngestedAt: time.Now(),
			}

			kr.logger.Debugw("parsed news event", &newsEvent)
			kr.consumeChan <- message
		}
	}
}
