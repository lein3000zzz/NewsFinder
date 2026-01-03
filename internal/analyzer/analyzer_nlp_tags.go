package analyzer

import (
	"NewsFinder/internal/analyzer/nlp"
	"NewsFinder/internal/analyzer/tagdetector"

	"go.uber.org/zap"
)

type CryptoAnalyzer struct {
	logger      *zap.SugaredLogger
	nlpAnalyzer nlp.AnalyzerNLP
	tagDetector tagdetector.TagDetector
}

func NewCryptoAnalyzer(logger *zap.SugaredLogger, nlpAnalyzer nlp.AnalyzerNLP, tagDetector tagdetector.TagDetector) *CryptoAnalyzer {
	return &CryptoAnalyzer{
		logger:      logger,
		nlpAnalyzer: nlpAnalyzer,
		tagDetector: tagDetector,
	}
}

func (ca *CryptoAnalyzer) Analyze(content string) (*AnalysisRes, error) {
	ca.logger.Infow("analyzing content", "content", content)

	nlpRes, err := ca.nlpAnalyzer.Analyze(content)
	if err != nil {
		ca.logger.Errorw("Error analyzing by nlp", "content", content, "error", err)
		return nil, err
	}

	tagsRes := ca.tagDetector.DetectTags(content)

	analysisRes := AnalysisRes{
		TagDetectorRes: tagsRes,
		NlpRes:         nlpRes,
	}

	return &analysisRes, nil
}
