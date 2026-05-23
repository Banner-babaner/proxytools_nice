package infrastructure

import (
	"fmt"
	"testing"
	"time"

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

func TestRemove_NoDeadlock(t *testing.T) {
	trie := NewIPTrie()
	trie.LoadLists(entity.ListsConfig{
		Blacklist: []string{"1.2.3.4", "5.6.7.8"},
		Whitelist: []string{"192.168.1.1"},
	})

	trie.Remove("1.2.3.4")

	_, found := trie.Search("1.2.3.4")
	assert.False(t, found)

	_, found = trie.Search("5.6.7.8")
	assert.True(t, found)

	_, found = trie.Search("192.168.1.1")
	assert.True(t, found)
}

func TestConcurrentLoadAndSearch(t *testing.T) {
	trie := NewIPTrie()
	trie.Insert("10.0.0.0/8", entity.Blacklist)

	done := make(chan bool, 20)

	for i := 0; i < 10; i++ {
		go func() {
			trie.LoadLists(entity.ListsConfig{
				Whitelist: []string{"192.168.1.0/24"},
				Blacklist: []string{"10.0.0.0/8"},
			})
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				trie.Search("10.1.1.1")
				trie.Search("192.168.1.1")
			}
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("deadlock detected")
		}
	}
}

func TestInsert_NoDeadlock(t *testing.T) {
	trie := NewIPTrie()

	for i := 0; i < 100; i++ {
		trie.Insert(fmt.Sprintf("10.0.%d.1", i), entity.Blacklist)
	}

	_, found := trie.Search("10.0.50.1")
	assert.True(t, found)
}

func TestInsertRange_NoDeadlock(t *testing.T) {
	trie := NewIPTrie()
	err := trie.InsertRange("192.168.1.1", "192.168.1.20", entity.Whitelist)
	assert.NoError(t, err)

	_, found := trie.Search("192.168.1.10")
	assert.True(t, found)
}

func TestGetLists_AfterLoad(t *testing.T) {
	trie := NewIPTrie()
	trie.LoadLists(entity.ListsConfig{
		Whitelist: []string{"192.168.1.1", "10.0.0.1"},
		Blacklist: []string{"1.2.3.4"},
	})

	lists := trie.GetLists()
	assert.Len(t, lists.Whitelist, 2)
	assert.Len(t, lists.Blacklist, 1)
	assert.Len(t, lists.Graylist, 0)
}