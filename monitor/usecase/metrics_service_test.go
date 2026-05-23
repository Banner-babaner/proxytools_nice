package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricsService(t *testing.T) {
	ms := NewMetricsService()
	assert.NotNil(t, ms)
}

func TestRecordRequest_Allowed(t *testing.T) {
	ms := NewMetricsService()
	ms.RecordRequest(true, 0.05, 1024, 2048)

	assert.Equal(t, int64(1), ms.totalRequests.Load())
	assert.Equal(t, int64(1), ms.allowedRequests.Load())
	assert.Equal(t, int64(0), ms.deniedRequests.Load())
	assert.Equal(t, int64(1024), ms.totalBytesUp.Load())
	assert.Equal(t, int64(2048), ms.totalBytesDown.Load())
}

func TestRecordRequest_Denied(t *testing.T) {
	ms := NewMetricsService()
	ms.RecordRequest(false, 0, 512, 0)

	assert.Equal(t, int64(1), ms.deniedRequests.Load())
}

func TestRecordCacheHit(t *testing.T) {
	ms := NewMetricsService()
	ms.RecordCacheHit()
	ms.RecordCacheHit()
	assert.Equal(t, int64(2), ms.cacheHits.Load())
}

func TestRecordCacheMiss(t *testing.T) {
	ms := NewMetricsService()
	ms.RecordCacheMiss()
	assert.Equal(t, int64(1), ms.cacheMisses.Load())
}

func TestRecordRateLimit(t *testing.T) {
	ms := NewMetricsService()
	ms.RecordRateLimit()
	assert.Equal(t, int64(1), ms.rateLimitedCount.Load())
}

func TestConnections(t *testing.T) {
	ms := NewMetricsService()
	ms.IncrementConnections()
	ms.IncrementConnections()
	assert.Equal(t, int64(2), ms.activeConns.Load())

	ms.DecrementConnections()
	assert.Equal(t, int64(1), ms.activeConns.Load())
}

func TestGetStats(t *testing.T) {
	ms := NewMetricsService()
	time.Sleep(100 * time.Millisecond)
	ms.RecordRequest(true, 0.05, 2048*1024, 4096*1024)

	stats := ms.GetStats()
	assert.Greater(t, stats.Uptime, 0.0)
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.GreaterOrEqual(t, stats.TotalUpMB, 1.0)
	assert.GreaterOrEqual(t, stats.TotalDownMB, 2.0)
}

func TestConcurrentRecording(t *testing.T) {
	ms := NewMetricsService()
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				ms.RecordRequest(true, 0.01, 100, 200)
				ms.RecordCacheHit()
				ms.IncrementConnections()
				ms.DecrementConnections()
			}
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	stats := ms.GetStats()
	assert.Equal(t, int64(5000), stats.TotalRequests)
	assert.Equal(t, int64(5000), stats.CacheHits)
	assert.Equal(t, int64(0), stats.ActiveConns)
}