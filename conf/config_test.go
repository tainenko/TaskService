package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func chdirWithConfig(t *testing.T, yaml string) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "conf"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "conf", "config.test.yaml"), []byte(yaml), 0o644))
	wd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	chdirWithConfig(t, "Server:\n  HttpPort: 8080\nDatabase:\n  Host: file-host\n  Password: file-pw\n")
	t.Setenv("TASK_DATABASE_PASSWORD", "env-pw")
	t.Setenv("TASK_SERVER_HTTPPORT", "9090")

	cfg, err := LoadConfig("test")
	require.NoError(t, err)
	assert.Equal(t, "env-pw", cfg.Database.Password)
	assert.Equal(t, "file-host", cfg.Database.Host)
	assert.Equal(t, 9090, cfg.Server.HttpPort)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	chdirWithConfig(t, "")
	_, err := LoadConfig("nope")
	assert.Error(t, err)
}
