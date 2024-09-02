package usecase

import (
	"context"
	"github.com/newmohr/example/internal/domain/entity"
	"github.com/newmohr/example/internal/domain/repository"
)

type LocationUseCase interface {
	GetLocations(ctx context.Context) ([]*entity.Location, error)
	UpdateLocations(ctx context.Context) error
}

type LocationUseCaseImpl struct {
	lc repository.LocationCache
	lr repository.LocationRepository
}

func NewLocationUseCase(
	lr repository.LocationRepository,
	lc repository.LocationCache,
) LocationUseCase {
	return &LocationUseCaseImpl{
		lr: lr,
		lc: lc,
	}
}

// GetLocations ロケーション情報を取得するメソッド
func (r *LocationUseCaseImpl) GetLocations(ctx context.Context) ([]*entity.Location, error) {
	locations, err := r.lc.FetchList(ctx)
	if err != nil {
		return nil, err
	}
	if len(locations) > 0 {
		return locations, nil
	}

	locations, err = r.lr.FetchList(ctx)
	if err != nil {
		return nil, err
	}

	err = r.lc.Update(ctx, locations)
	if err != nil {
		return nil, err
	}
	return locations, nil
}

// UpdateLocations ロケーション情報を更新するメソッド
func (r *LocationUseCaseImpl) UpdateLocations(ctx context.Context) error {
	locations, err := r.lr.FetchList(ctx)
	if err != nil {
		return err
	}

	err = r.lc.Update(ctx, locations)
	if err != nil {
		return err
	}
	return nil
}
