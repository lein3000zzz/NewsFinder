package tagdetector

import (
	"NewsFinder/internal/analyzer/tagdetector/externaldata"
	"regexp"
	"strings"

	"github.com/lein3000zzz/vault-config-manager/pkg/manager"
	"go.uber.org/zap"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

type MemoryTagDetector struct {
	logger        *zap.SugaredLogger
	sm            manager.SecretManager
	cacheEntities map[string]struct{}
}

func NewMemoryTagDetector(logger *zap.SugaredLogger, sm manager.SecretManager) *MemoryTagDetector {
	url, err := sm.GetSecretStringFromConfig("BINANCE_URL")
	if err != nil {
		logger.Fatalf("error getting secret string from config: %v", err)
	}
	set, err := externaldata.GetBinanceInfoSet(url)
	if err != nil {
		logger.Fatalf("error getting binance info: %v", err)
	}

	logger.Infow("loaded cache entities and initialized tag detector", "count", len(set))

	return &MemoryTagDetector{
		logger:        logger,
		cacheEntities: set,
		sm:            sm,
	}
}

func (md *MemoryTagDetector) DetectTags(content string) *DetectorRes {
	words := strings.Fields(nonAlphanumeric.ReplaceAllString(strings.ToUpper(content), " "))

	foundTags := make(map[string]struct{})

	for _, word := range words {
		if _, exists := md.cacheEntities[word]; exists {
			foundTags["#"+word] = struct{}{}
		}
	}

	var tags []string
	for tag := range foundTags {
		tags = append(tags, tag)
	}

	md.logger.Debugw("tags detector run", "tags", tags)

	res := DetectorRes{
		Tags: tags,
	}

	return &res
}
