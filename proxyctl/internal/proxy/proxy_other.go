//go:build !darwin && !windows

package proxy

import "errors"

// Get 在非 macOS/Windows 平台上无法读取系统代理。
func Get() (*Info, error) {
	return nil, errors.New("该系统代理命令仅在 macOS/Windows 上支持")
}

// Clear 在非 macOS/Windows 平台上不支持。
func Clear() error {
	return errors.New("该系统代理命令仅在 macOS/Windows 上支持")
}
