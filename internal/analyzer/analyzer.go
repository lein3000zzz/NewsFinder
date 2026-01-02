package analyzer

type AnalysisRes struct {
	DetectedTags []string
}

type Analyzer interface {
	Analyze(content string) (*AnalysisRes, error)
}
