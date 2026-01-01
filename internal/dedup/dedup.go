package dedup

import "NewsFinder/internal/pb/newsevent"

type HardDedupResult struct {
	Exists     bool
	Hash       string
	normalized string
}

type ManagerDedup interface {
	CheckExistsHard(event *newsevent.NewsEvent) (*HardDedupResult, error)
}
