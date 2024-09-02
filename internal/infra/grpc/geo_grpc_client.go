package grpc

import (
	"context"

	"github.com/newmohr/example/api/geo"
	"github.com/newmohr/example/internal/domain/entity"
	"github.com/newmohr/example/internal/domain/repository"
	"google.golang.org/grpc"
)

type GeoGrpc struct {
	client geo.GeoServerClient
}

func NewGeoGrpc(conn *grpc.ClientConn) repository.LocationRepository {
	client := geo.NewGeoServerClient(conn)
	return &GeoGrpc{client: client}
}

// FetchList gRPCサーバーからデータを取得するメソッド
func (gc *GeoGrpc) FetchList(ctx context.Context) ([]*entity.Location, error) {
	var allLocations []*entity.Location
	pageToken := ""

	for {
		req := &geo.ListLocationsRequest{
			PageToken: pageToken,
		}

		resp, err := gc.client.ListLocations(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, loc := range resp.Locations {
			allLocations = append(allLocations, &entity.Location{
				ID:   loc.Id,
				Name: loc.Name,
			})
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return allLocations, nil
}
