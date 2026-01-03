package datacombiner

import (
	"sync"

	"golang.org/x/sync/errgroup"
)

type DataSource struct {
	URL        string
	DataGetter func(url string) (map[string]struct{}, error)
}

func CombineData(dataGetters ...*DataSource) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	var mu sync.Mutex
	var eg errgroup.Group

	for _, dataGetter := range dataGetters {
		eg.Go(func() error {
			set, err := dataGetter.DataGetter(dataGetter.URL)
			if err != nil {
				return err
			}

			mu.Lock()
			defer mu.Unlock()

			for k := range set {
				result[k] = struct{}{}
			}

			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return result, err
}
