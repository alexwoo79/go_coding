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

// SnapshotSystem 在非 macOS/Windows 平台上不支持。
func SnapshotSystem() (SystemSnapshot, error) {
	return SystemSnapshot{}, errors.New("该系统代理快照仅在 macOS/Windows 上支持")
}

// RestoreSystem 在非 macOS/Windows 平台上不支持。
func RestoreSystem(SystemSnapshot) error {
	return errors.New("该系统代理快照仅在 macOS/Windows 上支持")
}

// ApplyProfile 在非 macOS/Windows 平台上不支持。
func ApplyProfile(*EndpointState, *EndpointState, *EndpointState, *PACState) error {
	return errors.New("该系统代理配置仅在 macOS/Windows 上支持")
}
