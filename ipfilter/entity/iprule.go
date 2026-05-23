package entity

type ListType int

const (
	Whitelist ListType = iota
	Blacklist
	Graylist
)

type AccessResult int

const (
	Allowed AccessResult = iota
	Denied
	CaptchaRequired
)

type ListsConfig struct {
	Whitelist []string
	Blacklist []string
	Graylist  []string
}

type FilterConfig struct {
	DefaultPolicy string
	Cache         CacheConfig
}

type CacheConfig struct {
	Enabled bool
	TTL     int
	MaxSize int
}