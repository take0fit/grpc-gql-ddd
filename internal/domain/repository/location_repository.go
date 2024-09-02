package repository

import (
	"context"

	"github.com/newmohr/example/internal/domain/entity"
)

//go:generate moq -out=../mock_repository/mock_$GOFILE -pkg=mock_repository . LocationRepository
type LocationRepository interface {
	FetchList(ctx context.Context) ([]*entity.Location, error)
}
