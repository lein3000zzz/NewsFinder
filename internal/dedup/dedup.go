package dedup

import "NewsFinder/internal/pb/news"

type HardDedupResult struct {
	Exists     bool
	Hash       string
	Normalized string
}

type SoftDedupRes struct {
	Exists bool
	Vector []float32
}

type ManagerDedup interface {
	CheckExistsHard(event *news.NewsEvent) (*HardDedupResult, error)
	CheckExistsSoft(hardRes *HardDedupResult) (*SoftDedupRes, error)
}
