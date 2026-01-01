package communicator

import (
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type KafkaCommunicator struct {
	logger      *zap.SugaredLogger
	kafkaClient *kgo.Client
}

func NewKafkaCommunicator(logger *zap.SugaredLogger, seeds []string, consumerGroup, topic string) *KafkaCommunicator {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		logger.Fatal(err)
	}

	return &KafkaCommunicator{
		logger:      logger,
		kafkaClient: client,
	}
}
