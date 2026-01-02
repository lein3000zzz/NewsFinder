package dedup

import (
	"NewsFinder/internal/datamanager"
	"NewsFinder/internal/pb/news"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"regexp"
	"strings"

	"github.com/clems4ever/all-minilm-l6-v2-go/all_minilm_l6_v2"
	"go.uber.org/zap"
	"golang.org/x/text/unicode/norm"
)

var (
	urlRegex   = regexp.MustCompile(`https?://\S+`)
	spaceRegex = regexp.MustCompile(`\s+`)
)

type PgDedupManager struct {
	logger         *zap.SugaredLogger
	dm             datamanager.DataManager
	embeddingModel *all_minilm_l6_v2.Model
}

func NewPgDedup(logger *zap.SugaredLogger, dm datamanager.DataManager) *PgDedupManager {
	model, err := all_minilm_l6_v2.NewModel(
		all_minilm_l6_v2.WithRuntimePath(os.Getenv("ONNX_PATH")))
	if err != nil {
		logger.Fatalw("Error creating model", "error", err)
	}

	return &PgDedupManager{
		logger:         logger,
		dm:             dm,
		embeddingModel: model,
	}
}

func (d *PgDedupManager) CheckExistsHard(event *news.NewsEvent) (*HardDedupResult, error) {
	normalized := d.normalize(event)
	hash := d.hashContent(normalized)

	exists, err := d.dm.LookupNewsByHash(hash)
	if err != nil {
		d.logger.Errorw("error looking up hash", "hash", hash, "error", err)
		return nil, err
	}

	d.logger.Debugw("found new hash", "hash", hash)

	res := &HardDedupResult{
		Exists:     exists,
		Hash:       hash,
		Normalized: normalized,
	}

	return res, nil
}

func (d *PgDedupManager) CheckExistsSoft(hardRes *HardDedupResult) (*SoftDedupRes, error) {
	vector, err := d.computeEmbedding(hardRes)
	if err != nil {
		d.logger.Errorw("error computing embedding", "error", err)
		return nil, err
	}

	exists, err := d.dm.LookupNewsByEmbedding(vector)
	if err != nil {
		d.logger.Errorw("error looking up embedding", "error", err)
		return nil, err
	}

	d.logger.Debugw("checked embedding", "exists", exists)

	res := &SoftDedupRes{
		Exists: exists,
		Vector: vector,
	}

	return res, nil
}

func (d *PgDedupManager) computeEmbedding(hardRes *HardDedupResult) ([]float32, error) {
	vector, err := d.embeddingModel.Compute(hardRes.Normalized, true)
	if err != nil {
		d.logger.Errorw("error computing embeddings", "error", err)
		return nil, err
	}

	return vector, nil
}

func (d *PgDedupManager) normalize(news *news.NewsEvent) string {
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
