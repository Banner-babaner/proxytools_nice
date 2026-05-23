package usecase

import (
	"testing"

	"github.com/Banner-babaner/proxytools_nice/ipfilter/entity"
	"github.com/Banner-babaner/proxytools_nice/ipfilter/mocks"
	"github.com/Banner-babaner/proxytools_nice/ipfilter/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckAccess_Whitelist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockCache.On("Get", "192.168.1.1").Return(entity.ListType(0), false, false)
	mockRepo.On("Search", "192.168.1.1").Return(entity.Whitelist, true)
	mockCache.On("Set", "192.168.1.1", entity.Whitelist, true).Return()

	fs := &FilterService{repo: mockRepo, cache: mockCache, defaultPolicy: "deny"}
	assert.Equal(t, entity.Allowed, fs.CheckAccess("192.168.1.1"))
}

func TestCheckAccess_Blacklist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockCache.On("Get", "10.0.0.1").Return(entity.ListType(0), false, false)
	mockRepo.On("Search", "10.0.0.1").Return(entity.Blacklist, true)
	mockCache.On("Set", "10.0.0.1", entity.Blacklist, true).Return()

	fs := &FilterService{repo: mockRepo, cache: mockCache, defaultPolicy: "allow"}
	assert.Equal(t, entity.Denied, fs.CheckAccess("10.0.0.1"))
}

func TestCheckAccess_Graylist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockCache.On("Get", "172.16.0.1").Return(entity.ListType(0), false, false)
	mockRepo.On("Search", "172.16.0.1").Return(entity.Graylist, true)
	mockCache.On("Set", "172.16.0.1", entity.Graylist, true).Return()

	fs := &FilterService{repo: mockRepo, cache: mockCache, defaultPolicy: "deny"}
	assert.Equal(t, entity.CaptchaRequired, fs.CheckAccess("172.16.0.1"))
}

func TestCheckAccess_DefaultDeny(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockCache.On("Get", "1.1.1.1").Return(entity.ListType(0), false, false)
	mockRepo.On("Search", "1.1.1.1").Return(entity.ListType(0), false)
	mockCache.On("Set", "1.1.1.1", entity.ListType(0), false).Return()

	fs := &FilterService{repo: mockRepo, cache: mockCache, defaultPolicy: "deny"}
	assert.Equal(t, entity.Denied, fs.CheckAccess("1.1.1.1"))
}

func TestCheckAccess_DefaultAllow(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockCache.On("Get", "1.1.1.1").Return(entity.ListType(0), false, false)
	mockRepo.On("Search", "1.1.1.1").Return(entity.ListType(0), false)
	mockCache.On("Set", "1.1.1.1", entity.ListType(0), false).Return()

	fs := &FilterService{repo: mockRepo, cache: mockCache, defaultPolicy: "allow"}
	assert.Equal(t, entity.Allowed, fs.CheckAccess("1.1.1.1"))
}

func TestCheckAccess_CacheHit(t *testing.T) {
	mockCache := new(mocks.IPCache)
	mockCache.On("Get", "10.0.0.1").Return(entity.Blacklist, true, true)

	fs := &FilterService{cache: mockCache, defaultPolicy: "deny"}
	assert.Equal(t, entity.Denied, fs.CheckAccess("10.0.0.1"))
}

func TestCheckAccess_CacheMissThenHit(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockCache.On("Get", "192.168.1.1").Return(entity.ListType(0), false, false).Once()
	mockRepo.On("Search", "192.168.1.1").Return(entity.Whitelist, true).Once()
	mockCache.On("Set", "192.168.1.1", entity.Whitelist, true).Return().Once()

	fs := &FilterService{repo: mockRepo, cache: mockCache, defaultPolicy: "deny"}
	assert.Equal(t, entity.Allowed, fs.CheckAccess("192.168.1.1"))

	mockCache.On("Get", "192.168.1.1").Return(entity.Whitelist, true, true).Once()
	assert.Equal(t, entity.Allowed, fs.CheckAccess("192.168.1.1"))

	mockRepo.AssertNumberOfCalls(t, "Search", 1)
}

func TestCheckAccess_NoCache(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockRepo.On("Search", "192.168.1.1").Return(entity.Whitelist, true)

	fs := &FilterService{repo: mockRepo, cache: nil, defaultPolicy: "deny"}
	assert.Equal(t, entity.Allowed, fs.CheckAccess("192.168.1.1"))
}

func TestAddIP_Blacklist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockRepo.On("Insert", "5.5.5.5", entity.Blacklist).Return(nil)

	fs := &FilterService{repo: mockRepo, defaultPolicy: "allow", lists: entity.ListsConfig{}}
	err := fs.AddIP("5.5.5.5", "blacklist")
	assert.NoError(t, err)
	assert.Contains(t, fs.lists.Blacklist, "5.5.5.5")
}

func TestAddIP_Whitelist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockRepo.On("Insert", "10.0.0.1", entity.Whitelist).Return(nil)

	fs := &FilterService{repo: mockRepo, defaultPolicy: "deny", lists: entity.ListsConfig{}}
	err := fs.AddIP("10.0.0.1", "whitelist")
	assert.NoError(t, err)
	assert.Contains(t, fs.lists.Whitelist, "10.0.0.1")
}

func TestAddIP_Graylist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockRepo.On("Insert", "172.16.0.1", entity.Graylist).Return(nil)

	fs := &FilterService{repo: mockRepo, defaultPolicy: "deny", lists: entity.ListsConfig{}}
	err := fs.AddIP("172.16.0.1", "graylist")
	assert.NoError(t, err)
	assert.Contains(t, fs.lists.Graylist, "172.16.0.1")
}

func TestAddIP_InvalidType(t *testing.T) {
	fs := &FilterService{}
	err := fs.AddIP("1.2.3.4", "purple")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown list type")
}

func TestAddIP_WithCache(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockRepo.On("Insert", "5.5.5.5", entity.Blacklist).Return(nil)
	mockCache.On("Remove", "5.5.5.5").Return()

	fs := &FilterService{repo: mockRepo, cache: mockCache, defaultPolicy: "allow", lists: entity.ListsConfig{}}
	err := fs.AddIP("5.5.5.5", "blacklist")
	assert.NoError(t, err)
	mockCache.AssertCalled(t, "Remove", "5.5.5.5")
}

func TestRemoveIP_Blacklist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

	fs := &FilterService{
		repo:    mockRepo,
		builder: func() repository.IPListRepository { return mockRepo },
		lists:   entity.ListsConfig{Blacklist: []string{"1.2.3.4"}},
	}

	fs.RemoveIP("1.2.3.4", "blacklist")
	assert.NotContains(t, fs.lists.Blacklist, "1.2.3.4")
}

func TestRemoveIP_Whitelist(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

	fs := &FilterService{
		repo:    mockRepo,
		builder: func() repository.IPListRepository { return mockRepo },
		lists:   entity.ListsConfig{Whitelist: []string{"192.168.1.1"}},
	}

	fs.RemoveIP("192.168.1.1", "whitelist")
	assert.Empty(t, fs.lists.Whitelist)
}

func TestRemoveIP_WithCache(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockCache := new(mocks.IPCache)

	mockRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	mockCache.On("Remove", "1.2.3.4").Return()

	fs := &FilterService{
		repo:    mockRepo,
		builder: func() repository.IPListRepository { return mockRepo },
		cache:   mockCache,
		lists:   entity.ListsConfig{Blacklist: []string{"1.2.3.4"}},
	}

	fs.RemoveIP("1.2.3.4", "blacklist")
	mockCache.AssertCalled(t, "Remove", "1.2.3.4")
}

func TestGetLists(t *testing.T) {
	fs := &FilterService{
		lists: entity.ListsConfig{
			Whitelist: []string{"192.168.1.1"},
			Blacklist: []string{"1.2.3.4"},
			Graylist:  []string{"172.16.0.1"},
		},
	}

	lists := fs.GetLists()
	assert.Len(t, lists.Whitelist, 1)
	assert.Len(t, lists.Blacklist, 1)
	assert.Len(t, lists.Graylist, 1)
}

func TestLoadLists(t *testing.T) {
	mockRepo := new(mocks.IPListRepository)
	mockRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

	fs := &FilterService{
		repo:    mockRepo,
		builder: func() repository.IPListRepository { return mockRepo },
	}

	lists := entity.ListsConfig{
		Whitelist: []string{"192.168.1.1"},
		Blacklist: []string{"1.2.3.4"},
	}

	fs.LoadLists(lists)
	assert.Equal(t, lists, fs.lists)
}

func TestDetermineAccess(t *testing.T) {
	fs := &FilterService{defaultPolicy: "deny"}

	assert.Equal(t, entity.Denied, fs.determineAccess(entity.Blacklist, true))
	assert.Equal(t, entity.Allowed, fs.determineAccess(entity.Whitelist, true))
	assert.Equal(t, entity.CaptchaRequired, fs.determineAccess(entity.Graylist, true))
	assert.Equal(t, entity.Denied, fs.determineAccess(0, false))
}

func TestDetermineAccess_DefaultAllow(t *testing.T) {
	fs := &FilterService{defaultPolicy: "allow"}
	assert.Equal(t, entity.Allowed, fs.determineAccess(0, false))
}