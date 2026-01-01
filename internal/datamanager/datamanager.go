package datamanager

type DataManager interface {
	LookupByHash(hash string) (bool, error)
}
