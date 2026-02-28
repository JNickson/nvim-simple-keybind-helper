package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/stretchr/testify/require"
)

func TestResolveConfigPath_FlagOverridesEnv(t *testing.T) {
	path, err := resolveConfigPath([]string{"--config", "./custom.json"}, "./from-env.json")
	require.NoError(t, err)
	require.Equal(t, "./custom.json", path)
}

func TestResolveConfigPath_UsesEnvWhenFlagMissing(t *testing.T) {
	path, err := resolveConfigPath(nil, " ./from-env.json ")
	require.NoError(t, err)
	require.Equal(t, "./from-env.json", path)
}

func TestResolveConfigPath_EmptyWhenUnset(t *testing.T) {
	path, err := resolveConfigPath(nil, "")
	require.NoError(t, err)
	require.Equal(t, "", path)
}

func TestResolveConfigPath_ReturnsErrorForUnknownFlag(t *testing.T) {
	_, err := resolveConfigPath([]string{"--unknown"}, "")
	require.Error(t, err)
}

func TestLoadConfig_AppliesDefaultsForMissingFields(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	content := `{
  "rows": [
    {"mode": "normal", "keybind": "gd", "action": "go to definition"}
  ]
}`

	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	cfg, err := loadConfig(configPath)
	require.NoError(t, err)

	require.Len(t, cfg.Columns, 3)
	require.Equal(t, 7, cfg.Height)
	require.Len(t, cfg.Rows, 1)
	require.Equal(t, "gd", cfg.Rows[0].Keybind)
}

func TestLoadConfig_ReturnsErrorForInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "bad.json")

	err := os.WriteFile(configPath, []byte("{not-json}"), 0o644)
	require.NoError(t, err)

	_, err = loadConfig(configPath)
	require.Error(t, err)
}

func TestConfigColumnsToTableColumns(t *testing.T) {
	columns := []columnConfig{{Title: "Mode", Width: 8}, {Title: "Keybind", Width: 16}}

	converted := configColumnsToTableColumns(columns)

	require.Equal(t, []table.Column{
		{Title: "Mode", Width: 8},
		{Title: "Keybind", Width: 16},
	}, converted)
}

func TestConfigRowsToTableRows(t *testing.T) {
	rows := []rowConfig{{Mode: "normal", Keybind: "gd", Action: "go to definition"}}

	converted := configRowsToTableRows(rows)

	require.Equal(t, []table.Row{{"normal", "gd", "go to definition"}}, converted)
}
