package tagdetector

import (
	"NewsFinder/internal/analyzer/tagdetector/externaldata"

	"github.com/lein3000zzz/vault-config-manager/pkg/manager"
	"go.uber.org/zap"
)

type MemoryDetector struct {
	logger        *zap.SugaredLogger
	sm            manager.SecretManager
	cacheEntities map[string]struct{}
}

func NewMemoryDetector(logger *zap.SugaredLogger, sm manager.SecretManager) *MemoryDetector {
	url, err := sm.GetSecretStringFromConfig("binance_url")
	if err != nil {
		logger.Fatalf("error getting secret string from config: %v", err)
	}
	set, err := externaldata.GetBinanceInfoSet(url)
	if err != nil {
		logger.Fatalf("error getting binance info: %v", err)
	}

	return &MemoryDetector{
		logger:        logger,
		cacheEntities: set,
		sm:            sm,
	}
}
