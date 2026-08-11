package toolproxy

import (
	"os"
	"path/filepath"
	"strings"
)

// cargoTool 管理 ~/.cargo/config.toml 的 [http] proxy。
type cargoTool struct{}

func (cargoTool) Name() string { return "cargo" }

func (cargoTool) Path() string {
	if h := os.Getenv("CARGO_HOME"); h != "" {
		return filepath.Join(h, "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cargo", "config.toml")
}

func (t cargoTool) Read() (map[string]*string, error) {
	out := map[string]*string{}
	text, err := readFileOrEmpty(t.Path())
	if err != nil {
		return nil, err
	}
	inHTTP := false
	for _, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "[") {
			sec := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			inHTTP = sec == "http"
			continue
		}
		if !inHTTP {
			continue
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok || strings.TrimSpace(key) != "proxy" {
			continue
		}
		v := strings.Trim(strings.TrimSpace(value), `"`)
		out["http.proxy"] = &v
	}
	return out, nil
}

func (t cargoTool) Write(vals map[string]*string) error {
	text, err := readFileOrEmpty(t.Path())
	if err != nil {
		return err
	}

	var out []string
	inHTTP := false
	httpExists := false
	flushHTTP := func() {
		if inHTTP {
			if v, ok := vals["http.proxy"]; ok && v != nil {
				out = append(out, `proxy = "`+*v+`"`)
			}
			inHTTP = false
		}
	}

	for _, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "[") {
			flushHTTP()
			sec := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			if sec == "http" {
				inHTTP = true
				httpExists = true
			}
			out = append(out, raw)
			continue
		}
		if inHTTP {
			key, _, ok := strings.Cut(trimmed, "=")
			if ok && strings.TrimSpace(key) == "proxy" {
				continue
			}
		}
		out = append(out, raw)
	}
	flushHTTP()

	if !httpExists {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, "[http]")
		if v, ok := vals["http.proxy"]; ok && v != nil {
			out = append(out, `proxy = "`+*v+`"`)
		}
	}
	return atomicWrite(t.Path(), []byte(strings.Join(out, "\n")+"\n"))
}

func (cargoTool) ProxyValues(proxyURL, _ string) map[string]*string {
	return map[string]*string{"http.proxy": &proxyURL}
}
