package main

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/app"
	"NewsFinder/tools/sqlc/nfsqlc"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

	queries := nfsqlc.New(pool)

	sources, err := queries.GetSources(ctx)
	if err != nil {
		logger.Fatalf("Error getting sources: %v", err)
	}
	logger.Infof("Found %d sources", len(sources))

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
