package dedup

import (
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/pb/newsevent"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/text/unicode/norm"
)

var (
	urlRegex   = regexp.MustCompile(`https?://\S+`)
	spaceRegex = regexp.MustCompile(`\s+`)
)

type PgDedupManager struct {
	logger *zap.SugaredLogger
	dm     datamanager.DataManager
}

func NewPgDedup(logger *zap.SugaredLogger, dm datamanager.DataManager) *PgDedupManager {
	return &PgDedupManager{
		logger: logger,
		dm:     dm,
	}
}

func (d *PgDedupManager) CheckExistsHard(event *newsevent.NewsEvent) (*HardDedupResult, error) {
	normalized := d.normalize(event)
	hash := d.hashContent(normalized)

	exists, err := d.dm.LookupByHash(hash)
	if err != nil {
		d.logger.Errorw("error looking up hash", "hash", hash, "error", err)
		return nil, err
	}

	d.logger.Debugw("found new hash", "hash", hash)

	res := &HardDedupResult{
		Exists:     exists,
		Hash:       hash,
		normalized: normalized,
	}

	return res, nil
}

func (d *PgDedupManager) normalize(news *newsevent.NewsEvent) string {
	normalized := news.Title + "\n" + news.Content

	normalized = norm.NFKC.String(normalized)
	normalized = strings.ToLower(normalized)

	normalized = urlRegex.ReplaceAllString(normalized, "")
	normalized = spaceRegex.ReplaceAllString(normalized, " ")

	normalized = strings.TrimSpace(normalized)
	return normalized
}

func (d *PgDedupManager) hashContent(normalized string) string {
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
