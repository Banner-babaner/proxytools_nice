package infrastructure

import (
	"fmt"
	"net"
	"sync"

	"github.com/Banner-babaner/proxytools_nice/ipfilter/entity"
	"github.com/Banner-babaner/proxytools_nice/ipfilter/repository"
)

type trieNode struct {
	Left    *trieNode
	Right   *trieNode
	Type    entity.ListType
	HasRule bool
}

type IPTrie struct {
	mu   sync.RWMutex
	root *trieNode
	list entity.ListsConfig
}

var _ repository.IPListRepository = (*IPTrie)(nil)

func NewIPTrie() *IPTrie {
	return &IPTrie{root: &trieNode{}}
}

func (t *IPTrie) Search(ipStr string) (entity.ListType, bool) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, false
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return 0, false
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	var lastMatch entity.ListType
	found := false

	for i := 0; i < 32; i++ {
		if node.HasRule {
			lastMatch = node.Type
			found = true
		}

		byteIdx := i / 8
		bitIdx := 7 - (i % 8)

		if ip4[byteIdx]&(1<<bitIdx) != 0 {
			if node.Right == nil {
				break
			}
			node = node.Right
		} else {
			if node.Left == nil {
				break
			}
			node = node.Left
		}
	}

	if node.HasRule {
		lastMatch = node.Type
		found = true
	}

	return lastMatch, found
}

func (t *IPTrie) Insert(cidr string, listType entity.ListType) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.insert(cidr, listType)
}


func (t *IPTrie) insert(cidr string, listType entity.ListType) error {
	// Убрать t.mu.Lock() отсюда
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		ip := net.ParseIP(cidr)
		if ip == nil {
			return fmt.Errorf("invalid ip or cidr: %s", cidr)
		}
		ipNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
	}

	ones, _ := ipNet.Mask.Size()
	ip := ipNet.IP.To4()

	node := t.root
	for i := 0; i < ones; i++ {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)

		if ip[byteIdx]&(1<<bitIdx) != 0 {
			if node.Right == nil {
				node.Right = &trieNode{}
			}
			node = node.Right
		} else {
			if node.Left == nil {
				node.Left = &trieNode{}
			}
			node = node.Left
		}
	}

	node.HasRule = true
	node.Type = listType
	return nil
}

func (t *IPTrie) InsertRange(startIP, endIP string, listType entity.ListType) error {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)
	if start == nil || end == nil {
		return fmt.Errorf("invalid ip range: %s-%s", startIP, endIP)
	}

	start4 := start.To4()
	end4 := end.To4()

	for ip := copyIP(start4); !ip.Equal(end4); incrementIP(ip) {
		t.insert(ip.String()+"/32", listType)
	}
	t.insert(end4.String()+"/32", listType)

	return nil
}

func (t *IPTrie) Remove(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.list.Whitelist = removeFromSlice(t.list.Whitelist, ip)
	t.list.Blacklist = removeFromSlice(t.list.Blacklist, ip)
	t.list.Graylist = removeFromSlice(t.list.Graylist, ip)
	t.rebuild()
}

func (t *IPTrie) GetLists() entity.ListsConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.list
}

func (t *IPTrie) LoadLists(lists entity.ListsConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.list = lists
	t.rebuild()
}

func (t *IPTrie) rebuild() {
	t.root = &trieNode{}

	for _, ip := range t.list.Whitelist {
		t.insert(ip, entity.Whitelist)
	}
	for _, ip := range t.list.Blacklist {
		t.insert(ip, entity.Blacklist)
	}
	for _, ip := range t.list.Graylist {
		t.insert(ip, entity.Graylist)
	}
}

func copyIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}