package toolproxy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// dockerTool 管理 ~/.docker/config.json 的 proxies.default.*。
type dockerTool struct{}

func (dockerTool) Name() string { return "docker" }

func (dockerTool) Path() string {
	if d := os.Getenv("DOCKER_CONFIG"); d != "" {
		return filepath.Join(d, "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".docker", "config.json")
}

var dockerKeys = []string{"proxies.default.httpProxy", "proxies.default.httpsProxy", "proxies.default.noProxy"}

func (t dockerTool) Read() (map[string]*string, error) {
	out := map[string]*string{}
	root, err := readDockerJSON(t.Path())
	if err != nil {
		return nil, err
	}
	if root == nil {
		return out, nil
	}
	proxies, _ := root["proxies"].(map[string]any)
	if proxies == nil {
		return out, nil
	}
	def, _ := proxies["default"].(map[string]any)
	if def == nil {
		return out, nil
	}
	for _, key := range []string{"httpProxy", "httpsProxy", "noProxy"} {
		if v, ok := def[key].(string); ok && v != "" {
			out["proxies.default."+key] = &v
		}
	}
	return out, nil
}

func (t dockerTool) Write(vals map[string]*string) error {
	root, err := readDockerJSON(t.Path())
	if err != nil {
		return err
	}
	if root == nil {
		root = map[string]any{}
	}
	proxies, _ := root["proxies"].(map[string]any)
	if proxies == nil {
		proxies = map[string]any{}
		root["proxies"] = proxies
	}
	def, _ := proxies["default"].(map[string]any)
	if def == nil {
		def = map[string]any{}
		proxies["default"] = def
	}
	for _, key := range []string{"httpProxy", "httpsProxy", "noProxy"} {
		k := "proxies.default." + key
		if v, ok := vals[k]; ok && v != nil {
			def[key] = *v
		} else {
			delete(def, key)
		}
	}
	// 清理空容器，避免留下 "proxies": {"default": {}}。
	if len(def) == 0 {
		delete(proxies, "default")
	}
	if len(proxies) == 0 {
		delete(root, "proxies")
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(t.Path(), append(data, '\n'))
}

func (dockerTool) ProxyValues(proxyURL, noProxy string) map[string]*string {
	return map[string]*string{
		"proxies.default.httpProxy":  &proxyURL,
		"proxies.default.httpsProxy": &proxyURL,
		"proxies.default.noProxy":    &noProxy,
	}
}

// readDockerJSON 读取 docker config.json；文件不存在时返回 (nil, nil)。
func readDockerJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", path, err)
	}
	return root, nil
}
