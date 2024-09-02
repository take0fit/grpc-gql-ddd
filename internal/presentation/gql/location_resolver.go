package gql

import (
	"context"
	"github.com/newmohr/example/api"
	"github.com/newmohr/example/internal/usecase"
)

type Resolver struct {
	lu usecase.LocationUseCase
}

func NewResolver(lu usecase.LocationUseCase) *Resolver {
	return &Resolver{lu: lu}
}

// Locations Location情報を取得する
func (r *queryResolver) Locations(ctx context.Context) ([]*api.Location, error) {
	locations, err := r.lu.GetLocations(ctx)
	if err != nil {
		return nil, err
	}

	gqlLocations := make([]*api.Location, len(locations))
	for i, loc := range locations {
		gqlLocations[i] = &api.Location{
			ID:   loc.ID,
			Name: loc.Name,
		}
	}
	return gqlLocations, nil
}

func (r *Resolver) Query() api.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
