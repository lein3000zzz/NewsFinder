package analyzer

import (
	"NewsFinder/internal/analyzer/nlp"
	"NewsFinder/internal/analyzer/tagdetector"
)

type AnalysisRes struct {
	tagDetectorRes *tagdetector.DetectorRes
	nlpRes         *nlp.CryptoBertRes
}

type Analyzer interface {
	Analyze(content string) (*AnalysisRes, error)
}
