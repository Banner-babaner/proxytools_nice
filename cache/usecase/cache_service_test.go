package usecase

import (
	"net/http"
	"testing"
	"time"

	"github.com/Banner-babaner/proxytools_nice/cache/entity"
	"github.com/Banner-babaner/proxytools_nice/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTTL_Enabled_GET(t *testing.T) {
	svc := &CacheService{enabled: true, defaultTTL: 60 * time.Second}
	assert.Equal(t, 60*time.Second, svc.GetTTL("GET", "/test", ""))
}

func TestGetTTL_Disabled(t *testing.T) {
	svc := &CacheService{enabled: false, defaultTTL: 60 * time.Second}
	assert.Equal(t, time.Duration(0), svc.GetTTL("GET", "/test", ""))
}

func TestGetTTL_NonGET(t *testing.T) {
	svc := &CacheService{enabled: true, defaultTTL: 60 * time.Second}
	assert.Equal(t, time.Duration(0), svc.GetTTL("POST", "/test", ""))
}

func TestGetTTL_CustomRule(t *testing.T) {
	svc := &CacheService{
		enabled:    true,
		defaultTTL: 60 * time.Second,
		rules: []entity.CacheRule{
			{Path: "/api/*", TTL: 0},
			{Path: "/static/*", TTL: 3600 * time.Second},
		},
	}
	assert.Equal(t, time.Duration(0), svc.GetTTL("GET", "/api/users", ""))
	assert.Equal(t, 3600*time.Second, svc.GetTTL("GET", "/static/app.js", ""))
}

func TestGenerateKey(t *testing.T) {
	svc := &CacheService{}
	k1 := svc.GenerateKey("GET", "/api/test")
	k2 := svc.GenerateKey("GET", "/api/test")
	k3 := svc.GenerateKey("POST", "/api/test")
	assert.Equal(t, k1, k2)
	assert.NotEqual(t, k1, k3)
}

func TestGet_Success(t *testing.T) {
	mockRepo := new(mocks.CacheRepository)
	entry := &entity.CacheEntry{StatusCode: 200, Body: []byte("data"), CreatedAt: time.Now(), TTL: 60 * time.Second}
	mockRepo.On("Get", "key1").Return(entry, true)

	svc := &CacheService{repo: mockRepo, enabled: true}
	result, ok := svc.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 200, result.StatusCode)
}

func TestGet_Expired(t *testing.T) {
	mockRepo := new(mocks.CacheRepository)
	entry := &entity.CacheEntry{StatusCode: 200, CreatedAt: time.Now().Add(-120 * time.Second), TTL: 60 * time.Second}
	mockRepo.On("Get", "key1").Return(entry, true)
	mockRepo.On("Delete", "key1").Return()

	svc := &CacheService{repo: mockRepo, enabled: true}
	_, ok := svc.Get("key1")
	assert.False(t, ok)
	mockRepo.AssertCalled(t, "Delete", "key1")
}

func TestGet_Disabled(t *testing.T) {
	svc := &CacheService{enabled: false}
	_, ok := svc.Get("key1")
	assert.False(t, ok)
}

func TestSet_Success(t *testing.T) {
	mockRepo := new(mocks.CacheRepository)
	mockRepo.On("Set", "key1", mock.AnythingOfType("*entity.CacheEntry")).Return()

	svc := &CacheService{repo: mockRepo, enabled: true}
	svc.Set("key1", 200, http.Header{}, []byte("data"), 60*time.Second, nil)
	mockRepo.AssertCalled(t, "Set", "key1", mock.Anything)
}

func TestSet_Disabled(t *testing.T) {
	mockRepo := new(mocks.CacheRepository)
	svc := &CacheService{repo: mockRepo, enabled: false}
	svc.Set("key1", 200, nil, []byte("data"), 60*time.Second, nil)
	mockRepo.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
}

func TestInvalidate_ByKey(t *testing.T) {
	mockRepo := new(mocks.CacheRepository)
	mockRepo.On("Delete", "key1").Return()

	svc := &CacheService{repo: mockRepo, enabled: true}
	count, err := svc.Invalidate(entity.InvalidateRequest{Key: "key1"})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestInvalidate_Clear(t *testing.T) {
	mockRepo := new(mocks.CacheRepository)
	mockRepo.On("Clear").Return()

	svc := &CacheService{repo: mockRepo, enabled: true}
	_, err := svc.Invalidate(entity.InvalidateRequest{Clear: true})
	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Clear")
}

func TestStats(t *testing.T) {
	mockRepo := new(mocks.CacheRepository)
	expected := entity.CacheStats{Entries: 10, SizeMB: 5.5, MaxSize: 100, Enabled: true}
	mockRepo.On("Stats").Return(expected)

	svc := &CacheService{repo: mockRepo, enabled: true}
	stats := svc.Stats()
	assert.Equal(t, 10, stats.Entries)
	assert.Equal(t, 5.5, stats.SizeMB)
}