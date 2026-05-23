package infrastructure

import (
	"testing"

	"github.com/Banner-babaner/proxytools_nice/ipfilter/entity"
	"github.com/stretchr/testify/assert"
)

func TestLoadLists_NoDeadlock(t *testing.T) {
	trie := NewIPTrie()
	lists := entity.ListsConfig{
		Whitelist: []string{"192.168.1.1", "10.0.0.0/8"},
		Blacklist: []string{"1.2.3.4", "5.6.7.0/24"},
		Graylist:  []string{"172.16.0.0/12"},
	}

	trie.LoadLists(lists)

	lt, found := trie.Search("192.168.1.1")
	assert.True(t, found)
	assert.Equal(t, entity.Whitelist, lt)

	lt, found = trie.Search("10.1.1.1")
	assert.True(t, found)
	assert.Equal(t, entity.Whitelist, lt)

	lt, found = trie.Search("1.2.3.4")
	assert.True(t, found)
	assert.Equal(t, entity.Blacklist, lt)

	lt, found = trie.Search("5.6.7.100")
	assert.True(t, found)
	assert.Equal(t, entity.Blacklist, lt)

	lt, found = trie.Search("172.20.1.1")
	assert.True(t, found)
	assert.Equal(t, entity.Graylist, lt)
}

func TestLoadLists_MultipleCalls(t *testing.T) {
	trie := NewIPTrie()

	trie.LoadLists(entity.ListsConfig{Blacklist: []string{"1.2.3.4"}})
	_, found := trie.Search("1.2.3.4")
	assert.True(t, found)

	trie.LoadLists(entity.ListsConfig{Whitelist: []string{"192.168.1.1"}})
	_, found = trie.Search("1.2.3.4")
	assert.False(t, found)
	_, found = trie.Search("192.168.1.1")
	assert.True(t, found)
}