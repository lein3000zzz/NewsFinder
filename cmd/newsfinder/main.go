package main

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/app"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lein3000zzz/vault-config-manager/pkg/manager"
	"go.uber.org/zap"
)

func main() {
	app.InitEnv()

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error initializing zap logger: %v", err)
	}

	logger := zapLogger.Sugar()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("PG_DSN"))
	if err != nil {
		logger.Fatalf("Error connecting to database: %v", err)
	}
	defer pool.Close()

	vaultAddress := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")

	secretManager, err := manager.NewSecretManager(vaultAddress, token, manager.DefaultBasePathData, manager.DefaultBasePathMetaData, logger)
	if err != nil {
		logger.Fatalf("Error initializing secret manager: %v", err)
	}

	//queries := nfsqlc.New(pool)

	inputText := "Binance Futures Will Launch USDⓈ-Margined COLLECTUSDT and MAGMAUSDT Perpetual Contract (2025-12-31)\n\n2025-12-31 13:15 (UTC): COLLECTUSDT Perpetual Contract with up to 20x leverage\n\n2025-12-31 13:30 (UTC): MAGMAUSDT Perpetual Contract with up to 20x leverage"

	an := analyzer.NewNLPAnalyzer(logger)
	res, err := an.Analyze(inputText)
	if err != nil {
		log.Fatalf("Error analyzing: %v", err)
	}

	logger.Infow("Result", "result", res)

	select {
	case <-time.After(15 * time.Second):
		fmt.Println("timeout")
	}
}
