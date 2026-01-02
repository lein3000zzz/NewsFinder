package app

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/analyzer/nlp"
	"NewsFinder/internal/analyzer/tagdetector"
	"NewsFinder/internal/communicator"
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/dedup"
	"NewsFinder/tools/sqlc/nfsqlc"
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/lein3000zzz/vault-config-manager/pkg/manager"
	"go.uber.org/zap"
)

func InitApp() *NewsFinder {
	InitEnv()

	logger := initLogger()

	sm := initSecretManager(logger)
	pgPool := initPgxPool()

	nlpAnalyzer := initNLPAnalyzer(logger)
	tagDetector := initTagDetector(logger, sm)
	an := initAnalyzer(logger, nlpAnalyzer, tagDetector)

	kafkaComm := initCommunicator(logger)
	dm := initDataManager(logger, pgPool)

	ded := initDedup(logger, dm)
	cfg := initAppConfig()

	return &NewsFinder{
		config:       cfg,
		logger:       logger,
		communicator: kafkaComm,
		dedup:        ded,
		dm:           dm,
		analyzer:     an,
	}
}

func initAppConfig() Config {
	return Config{
		ProduceMessages: true,
		SaveToDB:        true,
	}
}

func initDedup(logger *zap.SugaredLogger, dm datamanager.DataManager) dedup.ManagerDedup {
	return dedup.NewPgDedup(logger, dm)
}

func initDataManager(logger *zap.SugaredLogger, pool *pgxpool.Pool) datamanager.DataManager {
	queries := nfsqlc.New(pool)
	return datamanager.NewPgDataManager(logger, queries)
}

func initCommunicator(logger *zap.SugaredLogger) communicator.Communicator {
	addr := os.Getenv("KAFKA_ADDR")
	kafkaUser := os.Getenv("KAFKA_USERNAME")
	kafkaPassword := os.Getenv("KAFKA_PASSWORD")
	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	consumerTopic := os.Getenv("KAFKA_CONSUMER_TOPIC")

	consumerConfig := communicator.KafkaConfig{
		Seeds:         []string{addr},
		ConsumerGroup: consumerGroup,
		Topic:         consumerTopic,
		User:          kafkaUser,
		Password:      kafkaPassword,
	}

	producerTopic := os.Getenv("KAFKA_PRODUCER_TOPIC")
	producerConfig := communicator.KafkaConfig{
		Seeds:         []string{addr},
		ConsumerGroup: consumerGroup,
		Topic:         producerTopic,
		User:          kafkaUser,
		Password:      kafkaPassword,
	}

	return communicator.NewKafkaConsumer(logger, &consumerConfig, &producerConfig)
}

func initAnalyzer(logger *zap.SugaredLogger, nlpAnalyzer nlp.AnalyzerNLP, tagDetector tagdetector.TagDetector) analyzer.Analyzer {
	return analyzer.NewCryptoAnalyzer(logger, nlpAnalyzer, tagDetector)
}

func initTagDetector(logger *zap.SugaredLogger, sm manager.SecretManager) tagdetector.TagDetector {
	return tagdetector.NewMemoryTagDetector(logger, sm)
}

func initNLPAnalyzer(logger *zap.SugaredLogger) nlp.AnalyzerNLP {
	return nlp.NewNLPAnalyzerBert(logger)
}

func initPgxPool() *pgxpool.Pool {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("PG_DSN"))
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return pool
}

func initSecretManager(logger *zap.SugaredLogger) manager.SecretManager {
	sm, err := manager.NewSecretManager(os.Getenv("VAULT_ADDRESS"), os.Getenv("VAULT_TOKEN"), manager.DefaultBasePathData+"newsfinder/", manager.DefaultBasePathMetaData+"newsfinder/", logger)
	if err != nil {
		logger.Fatalf("Error initializing secret manager: %v", err)
	}

	keys := strings.Split(os.Getenv("VAULT_KEYS"), ",")

	sm.UnsealVault(keys)

	err = sm.ResetConfig()
	if err != nil {
		logger.Fatalw("failed to update config on start", "error", err)
		return nil
	}

	return sm
}

func initLogger() *zap.SugaredLogger {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error initializing zap logger: %v", err)
		return nil
	}

	logger := zapLogger.Sugar()
	return logger
}

func InitEnv() {
	err := godotenv.Load("main.env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}
