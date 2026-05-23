package usecase

import (
	"github.com/Banner-babaner/proxytools_nice/ratelimit/entity"
	"github.com/Banner-babaner/proxytools_nice/ratelimit/repository"
)

type RateLimitService struct {
	repo    repository.RateLimitRepository
	enabled bool
}

func NewRateLimitService(
	cfg entity.RateLimitConfig,
	repoBuilder func() repository.RateLimitRepository,
) *RateLimitService {
	return &RateLimitService{
		repo:    repoBuilder(),
		enabled: cfg.Enabled,
	}
}

func (s *RateLimitService) Allow(ip string) bool {
	if !s.enabled {
		return true
	}
	return s.repo.Allow(ip)
}

func (s *RateLimitService) IncrementConnections(ip string) bool {
	if !s.enabled {
		return true
	}
	return s.repo.IncrementConnections(ip)
}

func (s *RateLimitService) DecrementConnections(ip string) {
	if !s.enabled {
		return
	}
	s.repo.DecrementConnections(ip)
}

func (s *RateLimitService) GetStats(ip string) *entity.RateLimitStats {
	return s.repo.GetStats(ip)
}