package state

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/alexwoo79/go_coding/proxyctl/internal/git"
	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	want := &State{
		Version:   Version,
		Timestamp: time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC),
		System: proxy.SystemSnapshot{
			Services: []proxy.ServiceState{{
				Name: "Wi-Fi",
				HTTP: &proxy.EndpointState{Enabled: true, Host: "127.0.0.1", Port: "7890"},
				PAC:  &proxy.PACState{Enabled: true, URL: "http://127.0.0.1:8080/proxy.pac"},
			}},
		},
		Git: git.Snapshot{Values: map[string]string{
			"http.proxy": "http://127.0.0.1:7890",
		}},
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Fatal("Load 返回 nil")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load = %#v, want %#v", got, want)
	}
}

func TestLoadMissing(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	if got != nil {
		t.Errorf("Load missing = %#v, want nil", got)
	}
}

func TestRemoveMissing(t *testing.T) {
	if err := Remove(filepath.Join(t.TempDir(), "missing.json")); err != nil {
		t.Fatalf("Remove missing: %v", err)
	}
}

func TestDefaultPathEnvOverride(t *testing.T) {
	t.Setenv(EnvFile, "/tmp/proxyctl-state.json")
	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if got != "/tmp/proxyctl-state.json" {
		t.Errorf("DefaultPath = %q", got)
	}
}

func TestDefaultPathDefault(t *testing.T) {
	t.Setenv(EnvFile, "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("无法获取主目录: %v", err)
	}
	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	want := filepath.Join(home, ".config", "proxyctl", "state.json")
	if got != want {
		t.Errorf("DefaultPath = %q, want %q", got, want)
	}
}
