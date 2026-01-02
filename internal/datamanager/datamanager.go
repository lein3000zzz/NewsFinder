package datamanager

import "time"

const (
	dbTimeout = 3 * time.Second
)

type DataManager interface {
	LookupByHash(hash string) (bool, error)
	LookupByEmbedding(vector []float32) (bool, error)
}
