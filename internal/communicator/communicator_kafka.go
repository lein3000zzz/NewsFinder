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
	logger         *zap.SugaredLogger
	consumerClient *kgo.Client
	producerClient *kgo.Client
	consumeChan    chan *ConsumeMessage
	produceChan    chan *ProduceMessage
}

func NewKafkaConsumer(logger *zap.SugaredLogger, consumerConfig, producerConfig *KafkaConfig) *KafkaCommunicator {
	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(consumerConfig.Seeds...),
		kgo.ConsumerGroup(consumerConfig.ConsumerGroup),
		kgo.ConsumeTopics(consumerConfig.Topic),
		kgo.SASL(plain.Auth{
			User: consumerConfig.User,
			Pass: consumerConfig.Password,
		}.AsMechanism()),
	)
	if err != nil {
		logger.Fatal(err)
	}

	var producerClient *kgo.Client
	if producerConfig != nil {
		producerClient, err = kgo.NewClient(
			kgo.SeedBrokers(producerConfig.Seeds...),
			kgo.ConsumerGroup(producerConfig.ConsumerGroup),
			kgo.DefaultProduceTopic(producerConfig.Topic),
			kgo.SASL(plain.Auth{
				User: consumerConfig.User,
				Pass: consumerConfig.Password,
			}.AsMechanism()),
		)

		if err != nil {
			logger.Fatal(err)
		}
	}

	return &KafkaCommunicator{
		logger:         logger,
		consumerClient: consumerClient,
		producerClient: producerClient,
		consumeChan:    make(chan *ConsumeMessage, 500),
		produceChan:    make(chan *ProduceMessage, 100),
	}
}

func (kr *KafkaCommunicator) GetConsumeChan() <-chan *ConsumeMessage {
	return kr.consumeChan
}

func (kr *KafkaCommunicator) WriteToProduceChan(msg *ProduceMessage) {
	kr.produceChan <- msg
}

func (kr *KafkaCommunicator) StartTopicConsumer() {
	ctx := context.Background()

	kr.logger.Infow("starting new kafka topic consumer")
	for {
		kr.logger.Infow("waiting for new kafka topic consumer")
		fetches := kr.consumerClient.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			kr.logger.Errorw("Error polling fetches", "errors", errs)
			continue
		}

		kr.logger.Infow("new kafka message consumed")

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			var newsEvent news.NewsEvent
			err := proto.Unmarshal(record.Value, &newsEvent)
			if err != nil {
				kr.logger.Errorw("Error unmarshalling news event", "record", record)
				continue
			}

			kr.logger.Infow("new news event", "event", &newsEvent)

			message := &ConsumeMessage{
				Event:      &newsEvent,
				IngestedAt: time.Now(),
			}

			kr.logger.Infow("parsed news event", &newsEvent)
			kr.consumeChan <- message
		}
	}
}

func (kr *KafkaCommunicator) StartTopicProducer() {
	ctx := context.Background()

	for msg := range kr.produceChan {
		marshalled, err := proto.Marshal(msg.NewsAnalyzed)

		if err != nil {
			kr.logger.Errorw("Error marshalling news analyzed", "error", err)
			continue
		}

		record := &kgo.Record{
			Value: marshalled,
		}

		kr.producerClient.Produce(ctx, record, func(_ *kgo.Record, err error) {
			if err != nil {
				kr.logger.Errorw("Error producing message", "error", err)
			}
		})
	}
}
