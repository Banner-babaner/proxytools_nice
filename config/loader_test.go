package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_ValidConfig(t *testing.T) {
	yaml := `
server:
  port: 9090
  upstream: "http://example.com"
logging:
  level: "debug"
`
	f, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(yaml)
	assert.NoError(t, err)
	f.Close()

	cfg, err := Load(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestLoad_Defaults(t *testing.T) {
	yaml := `server: {}`
	f, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(f.Name())
	f.WriteString(yaml)
	f.Close()

	cfg, err := Load(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	f, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(f.Name())
	f.WriteString(`invalid: yaml: :`)
	f.Close()

	_, err := Load(f.Name())
	assert.Error(t, err)
}