package nlp

type AnalyzerNLP interface {
	Analyze(content string) (*CryptoBertRes, error)
}
