package profile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/alexwoo79/go_coding/proxyctl/internal/proxy"
)

func TestValidName(t *testing.T) {
	for _, ok := range []string{"clash", "office-1", "my.profile", "A_b.c"} {
		if !ValidName(ok) {
			t.Errorf("ValidName(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", ".", "..", "../x", "a/b", "a b", "a*b", "直接"} {
		if ValidName(bad) {
			t.Errorf("ValidName(%q) = true, want false", bad)
		}
	}
}

func TestSaveLoadListRemove(t *testing.T) {
	t.Setenv(EnvDir, t.TempDir())

	want := Profile{
		Name:  "clash",
		HTTP:  &proxy.EndpointState{Enabled: true, Host: "127.0.0.1", Port: "7890"},
		SOCKS: &proxy.EndpointState{Enabled: true, Host: "127.0.0.1", Port: "7891"},
	}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load("clash")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Fatal("Load 返回 nil")
	}
	if !reflect.DeepEqual(got, &want) {
		t.Errorf("Load = %#v, want %#v", got, &want)
	}

	list, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Name != "clash" {
		t.Errorf("List = %#v", list)
	}

	if err := Remove("clash"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(filepath.Join(t.TempDir(), "clash.json")); err == nil {
		t.Error("Remove 后文件仍存在")
	}
}

func TestLoadMissing(t *testing.T) {
	t.Setenv(EnvDir, t.TempDir())
	p, err := Load("nope")
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	if p != nil {
		t.Errorf("Load missing = %#v, want nil", p)
	}
}

func TestRemoveMissing(t *testing.T) {
	t.Setenv(EnvDir, t.TempDir())
	if err := Remove("nope"); err == nil {
		t.Fatal("Remove missing 应返回错误")
	}
}
