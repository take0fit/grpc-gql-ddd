package mock_repository

type MockLocationCache = LocationCacheMock

func NewMockLocationCache() *MockLocationCache {
	return &MockLocationCache{}
}

type MockLocationRepository = LocationRepositoryMock

func NewMockLocationRepository() *MockLocationRepository {
	return &MockLocationRepository{}
}
