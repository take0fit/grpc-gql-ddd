package repository

import (
	"context"
	"github.com/newmohr/example/internal/domain/entity"
)

//go:generate moq -out=../mock_repository/mock_$GOFILE -pkg=mock_repository . LocationCache
type LocationCache interface {
	FetchList(ctx context.Context) ([]*entity.Location, error)
	Update(ctx context.Context, locations []*entity.Location) error
}
