package analyzer

type AnalysisRes struct {
	Bearish float32
	Bullish float32
}

type Analyzer interface {
	Analyze(content string) string
}
