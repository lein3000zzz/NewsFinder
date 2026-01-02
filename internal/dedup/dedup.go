package dedup

import "NewsFinder/internal/pb/newsevent"

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
	CheckExistsHard(event *newsevent.NewsEvent) (*HardDedupResult, error)
	CheckExistsSoft(hardRes *HardDedupResult) (*SoftDedupRes, error)
}
