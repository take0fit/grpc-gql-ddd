package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/newmohr/example/internal/usecase"
)

type LocationScheduler struct {
	locationUseCase usecase.LocationUseCase
}

func NewLocationScheduler(locationUseCase usecase.LocationUseCase) *LocationScheduler {
	return &LocationScheduler{locationUseCase: locationUseCase}
}

// StartCacheUpdater キャッシュを定期的に更新するためのゴルーチンを起動します。
func (jc *LocationScheduler) StartCacheUpdater(ctx context.Context, updateInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Updating cache...")
				if err := jc.locationUseCase.UpdateLocations(ctx); err != nil {
					log.Printf("Failed to update cache: %v", err)
				}
			case <-ctx.Done():
				log.Println("Stopping cache updater due to context cancellation")
				return
			}
		}
	}()
}
