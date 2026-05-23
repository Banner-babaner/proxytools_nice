package config

import (
    "fmt"
    "os"

    "github.com/Banner-babaner/proxytools_nice/config/entity"
    "gopkg.in/yaml.v3"
)

func Load(path string) (*entity.Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }

    var cfg entity.Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    if cfg.Server.Port == 0 {
        cfg.Server.Port = 8080
    }
    if cfg.Server.Upstream == "" {
        cfg.Server.Upstream = "http://localhost:9000"
    }
    if cfg.Logging.Level == "" {
        cfg.Logging.Level = "info"
    }
    if cfg.Logging.Output == "" {
        cfg.Logging.Output = "stdout"
    }

    return &cfg, nil
}