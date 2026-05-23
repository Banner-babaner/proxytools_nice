package entity

type Config struct {
    Server    ServerConfig    `yaml:"server"`
    Auth      AuthConfig      `yaml:"auth"`
    IPFilter  IPFilterConfig  `yaml:"ip_filter"`
    RateLimit RateLimitConfig `yaml:"rate_limit"`
    Cache     CacheConfig     `yaml:"cache"`
    Logging   LoggingConfig   `yaml:"logging"`
}

type ServerConfig struct {
    Port     int    `yaml:"port"`
    Upstream string `yaml:"upstream"`
}

type AuthConfig struct {
    SecretKey string     `yaml:"secret_key"`
    TokenTTL  int        `yaml:"token_ttl"`
    Users     []AuthUser `yaml:"users"`
}

type AuthUser struct {
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Role     string `yaml:"role"`
}

type IPFilterConfig struct {
    DefaultPolicy string      `yaml:"default_policy"`
    Lists         ListsConfig `yaml:"lists"`
    Cache         struct {
        Enabled bool `yaml:"enabled"`
        TTL     int  `yaml:"ttl"`
        MaxSize int  `yaml:"max_size"`
    } `yaml:"cache"`
    AutoReload bool `yaml:"auto_reload"`
}

type ListsConfig struct {
    Whitelist []string `yaml:"whitelist"`
    Blacklist []string `yaml:"blacklist"`
    Graylist  []string `yaml:"graylist"`
}

type RateLimitConfig struct {
    Enabled bool              `yaml:"enabled"`
    Default RateLimitDefaults `yaml:"default"`
}

type RateLimitDefaults struct {
    RPS         int `yaml:"rps"`
    Connections int `yaml:"connections"`
}

type CacheConfig struct {
    Enabled    bool        `yaml:"enabled"`
    DefaultTTL int         `yaml:"default_ttl"`
    MaxSize    int         `yaml:"max_size"`
    Rules      []CacheRule `yaml:"rules"`
}

type CacheRule struct {
    Path   string `yaml:"path"`
    Domain string `yaml:"domain"`
    TTL    int    `yaml:"ttl"`
}

type LoggingConfig struct {
    Level  string `yaml:"level"`
    Output string `yaml:"output"`
}