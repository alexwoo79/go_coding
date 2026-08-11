package toolproxy

import (
	"os"
	"path/filepath"
)

// npmTool 管理 ~/.npmrc 的 proxy / https-proxy。
type npmTool struct{}

func (npmTool) Name() string { return "npm" }

func (npmTool) Path() string {
	if p := os.Getenv("npm_config_userconfig"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".npmrc")
}

func (t npmTool) Read() (map[string]*string, error) {
	return readINISection(t.Path(), "", []string{"proxy", "https-proxy"})
}

func (t npmTool) Write(vals map[string]*string) error {
	return writeINISection(t.Path(), "", []string{"proxy", "https-proxy"}, vals)
}

func (npmTool) ProxyValues(proxyURL, _ string) map[string]*string {
	return map[string]*string{"proxy": &proxyURL, "https-proxy": &proxyURL}
}
