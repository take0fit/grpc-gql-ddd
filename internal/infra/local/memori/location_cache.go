package memori

import (
	"context"
	"github.com/newmohr/example/internal/domain/entity"
	"github.com/newmohr/example/internal/domain/repository"
	"sync"
)

type LocationCache struct {
	cache []*entity.Location
	mutex sync.RWMutex
}

func NewLocationCache() repository.LocationCache {
	return &LocationCache{
		cache: []*entity.Location{},
	}
}

// FetchList オンメモリキャッシュからデータを取得するメソッド
func (lc *LocationCache) FetchList(ctx context.Context) ([]*entity.Location, error) {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return lc.cache, nil
	}
}

// Update オンメモリキャッシュを更新するメソッド
func (lc *LocationCache) Update(ctx context.Context, locations []*entity.Location) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		lc.cache = locations
		return nil
	}
}
