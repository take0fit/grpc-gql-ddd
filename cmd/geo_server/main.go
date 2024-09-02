package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/newmohr/example/api/geo"
	"google.golang.org/grpc"
)

type geoServer struct {
	geo.UnimplementedGeoServerServer
}

func (s *geoServer) ListLocations(ctx context.Context, req *geo.ListLocationsRequest) (*geo.ListLocationsResponse, error) {
	allLocations := []*geo.Location{
		{Id: "1", Name: "館山"},
		{Id: "2", Name: "鴨川"},
		{Id: "3", Name: "南房総市"},
		{Id: "4", Name: "鋸南町"},
		{Id: "5", Name: "白浜"},
		{Id: "6", Name: "千倉"},
		{Id: "7", Name: "和田町"},
		{Id: "8", Name: "丸山町"},
		{Id: "9", Name: "富浦"},
		{Id: "10", Name: "勝浦"},
		{Id: "11", Name: "御宿町"},
		{Id: "12", Name: "市原"},
		{Id: "13", Name: "木更津"},
		{Id: "14", Name: "君津"},
		{Id: "15", Name: "袖ケ浦"},
		{Id: "16", Name: "茂原"},
		{Id: "17", Name: "九十九里"},
		{Id: "18", Name: "いすみ"},
		{Id: "19", Name: "大原"},
		{Id: "20", Name: "大多喜町"},
	}

	pageSize := 3

	pageToken := req.GetPageToken()
	startIndex := 0
	if pageToken != "" {
		var err error
		startIndex, err = strconv.Atoi(pageToken)
		if err != nil || startIndex < 0 || startIndex >= len(allLocations) {
			return nil, fmt.Errorf("invalid page token")
		}
	}

	endIndex := startIndex + pageSize
	if endIndex > len(allLocations) {
		endIndex = len(allLocations)
	}

	locations := allLocations[startIndex:endIndex]

	var nextPageToken string
	if endIndex < len(allLocations) {
		nextPageToken = strconv.Itoa(endIndex)
	}

	return &geo.ListLocationsResponse{
		Locations:     locations,
		NextPageToken: nextPageToken,
	}, nil
}

func main() {
	grpcPort := getEnv("GRPC_SERVER_PORT", "50051")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	geo.RegisterGeoServerServer(s, &geoServer{})

	go func() {
		log.Printf("gRPC server is running on port: %s", grpcPort)
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down gRPC server...")
	s.GracefulStop()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
