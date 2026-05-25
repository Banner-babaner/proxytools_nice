package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	cacheInfra "github.com/Banner-babaner/proxytools_nice/cache/infrastructure"
	cacheUsecase "github.com/Banner-babaner/proxytools_nice/cache/usecase"
	filterEnt "github.com/Banner-babaner/proxytools_nice/ipfilter/entity"
	filterUsecase "github.com/Banner-babaner/proxytools_nice/ipfilter/usecase"
	"github.com/Banner-babaner/proxytools_nice/logger"
	monitorUsecase "github.com/Banner-babaner/proxytools_nice/monitor/usecase"
	ratelimitUsecase "github.com/Banner-babaner/proxytools_nice/ratelimit/usecase"
)

type ProxyHandler struct {
	reverseProxy *httputil.ReverseProxy
	ipFilter     *filterUsecase.FilterService
	rateLimiter  *ratelimitUsecase.RateLimitService
	cacheService *cacheUsecase.CacheService
	metrics      *monitorUsecase.MetricsService
}

func NewProxyHandler(
	upstreamURL string,
	ipFilter *filterUsecase.FilterService,
	rateLimiter *ratelimitUsecase.RateLimitService,
	cacheService *cacheUsecase.CacheService,
	metrics *monitorUsecase.MetricsService,
) (*ProxyHandler, error) {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}

	return &ProxyHandler{
		reverseProxy: proxy,
		ipFilter:     ipFilter,
		rateLimiter:  rateLimiter,
		cacheService: cacheService,
		metrics:      metrics,
	}, nil
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.metrics.IncrementConnections()
	// defer ph.metrics.DecrementConnections()

	startTime := time.Now()
	clientIP := r.RemoteAddr

	if host, _, err := net.SplitHostPort(clientIP); err == nil {
		clientIP = host
	}

	logger.Info(fmt.Sprintf("Checking access for: %s", clientIP))

	access := ph.ipFilter.CheckAccess(clientIP)

	switch access {
	case filterEnt.Allowed:
		logger.Info("Access: allowed")
	case filterEnt.Denied:
		logger.Info("Access: denied")
	case filterEnt.CaptchaRequired:
		logger.Info("Access: captcha")
	}

	if access == filterEnt.Denied {
		ph.metrics.RecordRequest(false, 0, 0, 0)
		http.Error(w, "Access denied", http.StatusForbidden)
		logger.Warn("Access denied: " + clientIP)
		return
	}

	if !ph.rateLimiter.Allow(clientIP) {
		ph.metrics.RecordRateLimit()
		ph.metrics.RecordRequest(false, time.Since(startTime).Seconds(), 0, 0)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	cacheKey := ph.cacheService.GenerateKey(r.Method, r.URL.String())
	ttl := ph.cacheService.GetTTL(r.Method, r.URL.Path, r.Host)

	if ttl > 0 {
		if entry, ok := ph.cacheService.Get(cacheKey); ok {
			ph.metrics.RecordCacheHit()
			ph.metrics.RecordRequest(true, time.Since(startTime).Seconds(), 0, int64(len(entry.Body)))

			for k, v := range entry.Headers {
				for _, val := range v {
					w.Header().Add(k, val)
				}
			}
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(entry.StatusCode)
			w.Write(entry.Body)
			return
		}
		ph.metrics.RecordCacheMiss()
	}

	crw := cacheInfra.NewCacheResponseWriter(w)
	ph.reverseProxy.ServeHTTP(crw, r)

	logger.Info(fmt.Sprintf("Response: status=%d, body=%d bytes", crw.StatusCode(), len(crw.Body())))

	duration := time.Since(startTime).Seconds()

	uploadSize := r.ContentLength
	if uploadSize < 0 {
		uploadSize = 0
	}

	ph.metrics.RecordRequest(true, duration, uploadSize, int64(len(crw.Body())))

	if ttl > 0 && crw.StatusCode() >= 200 && crw.StatusCode() < 300 {
		ph.cacheService.Set(cacheKey, crw.StatusCode(), w.Header(), crw.Body(), ttl, nil)
	}
}