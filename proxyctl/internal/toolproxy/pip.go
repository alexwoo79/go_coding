package toolproxy

import (
	"os"
	"path/filepath"
	"runtime"
)

// pipTool 管理 pip.conf 的 [global] proxy。
type pipTool struct{}

func (pipTool) Name() string { return "pip" }

func (pipTool) Path() string {
	if p := os.Getenv("PIP_CONFIG_FILE"); p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "pip", "pip.ini")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pip", "pip.conf")
}

func (t pipTool) Read() (map[string]*string, error) {
	raw, err := readINISection(t.Path(), "global", []string{"proxy"})
	if err != nil {
		return nil, err
	}
	out := map[string]*string{}
	if v, ok := raw["proxy"]; ok {
		out["global.proxy"] = v
	}
	return out, nil
}

func (t pipTool) Write(vals map[string]*string) error {
	ini := map[string]*string{}
	if v, ok := vals["global.proxy"]; ok {
		ini["proxy"] = v
	}
	return writeINISection(t.Path(), "global", []string{"proxy"}, ini)
}

func (pipTool) ProxyValues(proxyURL, _ string) map[string]*string {
	return map[string]*string{"global.proxy": &proxyURL}
}
