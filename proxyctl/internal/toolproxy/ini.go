package toolproxy

import (
	"strings"
)

// readINISection 读取 INI 风格（key=value）文件中指定 section 的 key。
// section 为空表示顶层（无 section 文件，如 .npmrc）。
func readINISection(path, section string, keys []string) (map[string]*string, error) {
	out := map[string]*string{}
	text, err := readFileOrEmpty(path)
	if err != nil {
		return nil, err
	}
	inSection := section == ""
	for _, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sec := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			inSection = section == "" || sec == section
			continue
		}
		if !inSection || trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if containsKey(keys, key) {
			v := strings.TrimSpace(value)
			out[key] = &v
		}
	}
	return out, nil
}

// writeINISection 写回 INI 文件：保留其他内容，按 keys 管理目标 section 的配置项。
// vals 中缺失或为 nil 的 key 会被移除。
func writeINISection(path, section string, keys []string, vals map[string]*string) error {
	text, err := readFileOrEmpty(path)
	if err != nil {
		return err
	}

	var out []string
	inSection := section == ""
	sectionFound := section == ""
	flushSection := func() {
		if inSection {
			for _, k := range keys {
				if v, ok := vals[k]; ok && v != nil {
					out = append(out, k+"="+*v)
				}
			}
		}
		inSection = false
	}

	for _, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			flushSection()
			sec := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			inSection = section == "" || sec == section
			if inSection {
				sectionFound = true
			}
			out = append(out, raw)
			continue
		}
		if inSection {
			key, _, ok := strings.Cut(trimmed, "=")
			if ok && containsKey(keys, strings.TrimSpace(key)) {
				continue // 删除旧值，由 flushSection 统一输出
			}
		}
		out = append(out, raw)
	}
	flushSection()

	if !sectionFound && section != "" {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, "["+section+"]")
		for _, k := range keys {
			if v, ok := vals[k]; ok && v != nil {
				out = append(out, k+"="+*v)
			}
		}
	}
	// 去掉末尾多余空行（新建空文件时首行可能为空）。
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	return atomicWrite(path, []byte(strings.Join(out, "\n")+"\n"))
}
