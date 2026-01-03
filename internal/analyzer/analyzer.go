package analyzer

import (
	"NewsFinder/internal/analyzer/nlp"
	"NewsFinder/internal/analyzer/tagdetector"
)

type AnalysisRes struct {
	TagDetectorRes *tagdetector.DetectorRes `json:"tag_detector"`
	NlpRes         *nlp.CryptoBertRes       `json:"nlp_crypto_bert"`
}

type Analyzer interface {
	Analyze(content string) (*AnalysisRes, error)
}
