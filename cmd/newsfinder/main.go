package main

import (
	"NewsFinder/internal/analyzer"
	"NewsFinder/internal/app"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
)

func main() {
	app.InitEnv()

	inputText := "Binance Futures Will Launch USDⓈ-Margined COLLECTUSDT and MAGMAUSDT Perpetual Contract (2025-12-31)\n\n2025-12-31 13:15 (UTC): COLLECTUSDT Perpetual Contract with up to 20x leverage\n\n2025-12-31 13:30 (UTC): MAGMAUSDT Perpetual Contract with up to 20x leverage"
	// Prepare tensors

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error initializing zap logger: %v", err)
	}

	logger := zapLogger.Sugar()

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
