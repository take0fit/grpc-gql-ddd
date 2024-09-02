package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/newmohr/example/internal/domain/entity"
	"github.com/newmohr/example/internal/domain/mock_repository"
)

type mockLocationUseCase struct {
	*LocationUseCaseImpl
	lc *mock_repository.MockLocationCache
	lr *mock_repository.MockLocationRepository
}

func newMockLocationUseCase() *mockLocationUseCase {
	lc := &mock_repository.MockLocationCache{}
	lr := &mock_repository.MockLocationRepository{}

	uc := NewLocationUseCase(lr, lc)

	return &mockLocationUseCase{
		LocationUseCaseImpl: uc.(*LocationUseCaseImpl),
		lc:                  lc,
		lr:                  lr,
	}
}

func TestLocationUseCase_GetLocations(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockLocationUseCase)
		expectedErr error
		want        []*entity.Location
	}{
		{
			name: "Success - Fetch from cache",
			mockSetup: func(m *mockLocationUseCase) {
				m.lc.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return []*entity.Location{
						{ID: "1", Name: "Tokyo"},
					}, nil
				}
			},
			want: []*entity.Location{
				{ID: "1", Name: "Tokyo"},
			},
			expectedErr: nil,
		},
		{
			name: "Success - Fetch from gRPC",
			mockSetup: func(m *mockLocationUseCase) {
				m.lc.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return nil, nil
				}
				m.lr.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return []*entity.Location{
						{ID: "1", Name: "Tokyo"},
					}, nil
				}
				m.lc.UpdateFunc = func(ctx context.Context, locations []*entity.Location) error {
					return nil
				}
			},
			want: []*entity.Location{
				{ID: "1", Name: "Tokyo"},
			},
			expectedErr: nil,
		},
		{
			name: "Cache Fetch Error",
			mockSetup: func(m *mockLocationUseCase) {
				m.lc.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return nil, errors.New("cache error")
				}
				m.lr.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return nil, nil
				}
			},
			want:        nil,
			expectedErr: errors.New("cache error"),
		},
		{
			name: "Repository error",
			mockSetup: func(m *mockLocationUseCase) {
				m.lc.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return nil, nil
				}
				m.lr.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return nil, errors.New("repository error")
				}
			},
			want:        nil,
			expectedErr: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockLocationUseCase()
			tt.mockSetup(m)

			output, err := m.GetLocations(context.Background())
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				} else if diff := cmp.Diff(output, tt.want); diff != "" {
					t.Errorf("GetLocations() mismatch (-got +want):\n%s", diff)
				}
			} else {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestLocationUseCase_UpdateLocations(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockLocationUseCase)
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func(m *mockLocationUseCase) {
				m.lr.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return []*entity.Location{
						{ID: "1", Name: "Tokyo"},
					}, nil
				}
				m.lc.UpdateFunc = func(ctx context.Context, locations []*entity.Location) error {
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name: "Repository error",
			mockSetup: func(m *mockLocationUseCase) {
				m.lr.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return nil, errors.New("repository error")
				}
			},
			expectedErr: errors.New("repository error"),
		},
		{
			name: "Cache Update error",
			mockSetup: func(m *mockLocationUseCase) {
				m.lr.FetchListFunc = func(ctx context.Context) ([]*entity.Location, error) {
					return []*entity.Location{
						{ID: "1", Name: "Tokyo"},
					}, nil
				}
				m.lc.UpdateFunc = func(ctx context.Context, locations []*entity.Location) error {
					return errors.New("cache update error")
				}
			},
			expectedErr: errors.New("cache update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockLocationUseCase()
			tt.mockSetup(m)

			err := m.UpdateLocations(context.Background())
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}
