package usecase

import (
	"testing"

	"github.com/Banner-babaner/proxytools_nice/ratelimit/entity"
	"github.com/Banner-babaner/proxytools_nice/ratelimit/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAllow_Enabled_True(t *testing.T) {
	mockRepo := new(mocks.RateLimitRepository)
	mockRepo.On("Allow", "192.168.1.1").Return(true)

	svc := &RateLimitService{repo: mockRepo, enabled: true}
	assert.True(t, svc.Allow("192.168.1.1"))
}

func TestAllow_Exceeded(t *testing.T) {
	mockRepo := new(mocks.RateLimitRepository)
	mockRepo.On("Allow", "10.0.0.1").Return(false)

	svc := &RateLimitService{repo: mockRepo, enabled: true}
	assert.False(t, svc.Allow("10.0.0.1"))
}

func TestAllow_Disabled(t *testing.T) {
	mockRepo := new(mocks.RateLimitRepository)
	svc := &RateLimitService{repo: mockRepo, enabled: false}
	assert.True(t, svc.Allow("192.168.1.1"))
}

func TestIncrementConnections_Enabled(t *testing.T) {
	mockRepo := new(mocks.RateLimitRepository)
	mockRepo.On("IncrementConnections", "192.168.1.1").Return(true)

	svc := &RateLimitService{repo: mockRepo, enabled: true}
	assert.True(t, svc.IncrementConnections("192.168.1.1"))
}

func TestIncrementConnections_Disabled(t *testing.T) {
	svc := &RateLimitService{repo: nil, enabled: false}
	assert.True(t, svc.IncrementConnections("192.168.1.1"))
}

func TestDecrementConnections_Enabled(t *testing.T) {
	mockRepo := new(mocks.RateLimitRepository)
	mockRepo.On("DecrementConnections", "192.168.1.1").Return()

	svc := &RateLimitService{repo: mockRepo, enabled: true}
	svc.DecrementConnections("192.168.1.1")
	mockRepo.AssertCalled(t, "DecrementConnections", "192.168.1.1")
}

func TestGetStats_Found(t *testing.T) {
	mockRepo := new(mocks.RateLimitRepository)
	expected := &entity.RateLimitStats{Tokens: 5.0, RPS: 10.0, Connections: 2, MaxConns: 50}
	mockRepo.On("GetStats", "192.168.1.1").Return(expected)

	svc := &RateLimitService{repo: mockRepo, enabled: true}
	stats := svc.GetStats("192.168.1.1")
	assert.NotNil(t, stats)
	assert.Equal(t, 5.0, stats.Tokens)
	assert.Equal(t, 2, stats.Connections)
}

func TestGetStats_NotFound(t *testing.T) {
	mockRepo := new(mocks.RateLimitRepository)
	mockRepo.On("GetStats", "unknown").Return(nil)

	svc := &RateLimitService{repo: mockRepo, enabled: true}
	assert.Nil(t, svc.GetStats("unknown"))
}