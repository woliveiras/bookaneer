package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.BindAddress)
	assert.Equal(t, "", cfg.URLBase)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "forms", cfg.AuthMethod)
	assert.Equal(t, "/library", cfg.LibraryDir)
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir, filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, dir, cfg.DataDir)
}

func TestLoad_FromYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	yaml := "port: 8080\nbindAddress: \"127.0.0.1\"\nlogLevel: \"debug\"\nlibraryDir: \"/books\"\n"
	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0644))

	cfg, err := config.Load(dir, cfgPath)
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "127.0.0.1", cfg.BindAddress)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "/books", cfg.LibraryDir)
}

func TestLoad_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BOOKANEER_PORT", "3000")
	t.Setenv("BOOKANEER_BIND_ADDRESS", "localhost")
	t.Setenv("BOOKANEER_LOG_LEVEL", "warn")
	t.Setenv("BOOKANEER_LIBRARY_DIR", "/my/books")

	cfg, err := config.Load(dir, filepath.Join(dir, "nonexistent.yaml"))
	require.NoError(t, err)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "localhost", cfg.BindAddress)
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, "/my/books", cfg.LibraryDir)
}

func TestLoad_EnvDataDir(t *testing.T) {
	dir := t.TempDir()
	customData := filepath.Join(dir, "custom-data")
	t.Setenv("BOOKANEER_DATA_DIR", customData)

	cfg, err := config.Load("", "")
	require.NoError(t, err)
	assert.Equal(t, customData, cfg.DataDir)
}

func TestDatabasePath(t *testing.T) {
	cfg := &config.Config{DataDir: "/var/data"}
	assert.Equal(t, "/var/data/bookaneer.db", cfg.DatabasePath())
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(":::invalid"), 0644))

	_, err := config.Load(dir, cfgPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse config")
}

func TestLoad_InvalidPort(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BOOKANEER_PORT", "notanumber")

	cfg, err := config.Load(dir, filepath.Join(dir, "nonexistent.yaml"))
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Port)
}

func TestLoad_EmptyDataDir_DefaultsToData(t *testing.T) {
	// Ensure no env var is set
	t.Setenv("BOOKANEER_DATA_DIR", "")

	// Use temp as working dir so "./data" resolves safely
	oldWd, _ := os.Getwd()
	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	cfg, err := config.Load("", "")
	require.NoError(t, err)
	assert.Equal(t, "./data", cfg.DataDir)
}

func TestLoad_EmptyConfigPath_DefaultsToDataDir(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir, "")
	require.NoError(t, err)
	assert.Equal(t, dir, cfg.DataDir)
	assert.Equal(t, 9090, cfg.Port)
}
