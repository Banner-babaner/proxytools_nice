package usecase

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Banner-babaner/proxytools_nice/ipfilter/entity"
	"github.com/Banner-babaner/proxytools_nice/ipfilter/repository"
	"github.com/Banner-babaner/proxytools_nice/logger"
)

type FilterService struct {
	mu            sync.RWMutex
	builder       func() repository.IPListRepository
	repo          repository.IPListRepository
	cache         repository.IPCache
	defaultPolicy string
	lists         entity.ListsConfig
}

func NewFilterService(
	cfg entity.FilterConfig,
	repoBuilder func() repository.IPListRepository,
	cacheBuilder func(maxSize int, ttlSeconds int) (repository.IPCache, error),
) (*FilterService, error) {
	fs := &FilterService{
		builder:       repoBuilder,
		repo:          repoBuilder(),
		defaultPolicy: cfg.DefaultPolicy,
	}

	if cfg.Cache.Enabled && cacheBuilder != nil {
		cache, err := cacheBuilder(cfg.Cache.MaxSize, cfg.Cache.TTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create ip cache: %w", err)
		}
		fs.cache = cache
	}

	return fs, nil
}

func (fs *FilterService) CheckAccess(ip string) entity.AccessResult {
	if fs.cache != nil {
		if listType, hasRule, found := fs.cache.Get(ip); found {
			return fs.determineAccess(listType, hasRule)
		}
	}

	fs.mu.RLock()
	listType, hasRule := fs.repo.Search(ip)
	fs.mu.RUnlock()

	if fs.cache != nil {
		fs.cache.Set(ip, listType, hasRule)
	}

	return fs.determineAccess(listType, hasRule)
}

func (fs *FilterService) determineAccess(listType entity.ListType, hasRule bool) entity.AccessResult {
	if hasRule {
		switch listType {
		case entity.Blacklist:
			return entity.Denied
		case entity.Whitelist:
			return entity.Allowed
		case entity.Graylist:
			return entity.CaptchaRequired
		}
	}

	if fs.defaultPolicy == "allow" {
		return entity.Allowed
	}
	return entity.Denied
}

func (fs *FilterService) AddIP(ip string, listType string) error {
	var lt entity.ListType
	switch listType {
	case "whitelist":
		lt = entity.Whitelist
		fs.lists.Whitelist = append(fs.lists.Whitelist, ip)
	case "blacklist":
		lt = entity.Blacklist
		fs.lists.Blacklist = append(fs.lists.Blacklist, ip)
	case "graylist":
		lt = entity.Graylist
		fs.lists.Graylist = append(fs.lists.Graylist, ip)
	default:
		return fmt.Errorf("unknown list type: %s", listType)
	}

	var err error
	if strings.Contains(ip, "-") {
		parts := strings.Split(ip, "-")
		if len(parts) == 2 {
			err = fs.repo.InsertRange(parts[0], parts[1], lt)
		} else {
			err = fmt.Errorf("invalid range format: %s", ip)
		}
	} else {
		err = fs.repo.Insert(ip, lt)
	}

	if err != nil {
		return fmt.Errorf("failed to insert ip %s: %w", ip, err)
	}

	if fs.cache != nil {
		fs.cache.Remove(ip)
	}

	return nil
}
func (fs *FilterService) RemoveIP(ip string, listType string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	switch listType {
	case "whitelist":
		fs.lists.Whitelist = removeFromSlice(fs.lists.Whitelist, ip)
	case "blacklist":
		fs.lists.Blacklist = removeFromSlice(fs.lists.Blacklist, ip)
	case "graylist":
		fs.lists.Graylist = removeFromSlice(fs.lists.Graylist, ip)
	}

	fs.loadListsNoLock(fs.lists)

	if fs.cache != nil {
		fs.cache.Remove(ip)
	}

	logger.Info(fmt.Sprintf("IP %s removed from %s", ip, listType))
}

func (fs *FilterService) GetLists() entity.ListsConfig {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.lists
}

func (fs *FilterService) LoadLists(lists entity.ListsConfig) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.lists = lists
	fs.loadListsNoLock(lists)
}

func (fs *FilterService) loadListsNoLock(lists entity.ListsConfig) {
	fs.repo = fs.builder()

	for _, ip := range lists.Blacklist {
		fs.repo.Insert(ip, entity.Blacklist)
	}
	for _, ip := range lists.Whitelist {
		fs.repo.Insert(ip, entity.Whitelist)
	}
	for _, ip := range lists.Graylist {
		fs.repo.Insert(ip, entity.Graylist)
	}

	logger.Info(fmt.Sprintf("IP lists loaded: blacklist=%d whitelist=%d graylist=%d",
		len(lists.Blacklist), len(lists.Whitelist), len(lists.Graylist)))
}

func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}