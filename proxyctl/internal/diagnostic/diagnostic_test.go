package diagnostic

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type fakeChecker struct {
	name string
	res  CheckResult
}

func (f fakeChecker) Name() string { return f.name }

func (f fakeChecker) Check(context.Context) CheckResult { return f.res }

func TestRunAll(t *testing.T) {
	boom := errors.New("boom")
	checkers := []Checker{
		fakeChecker{name: "a", res: CheckResult{Name: "a", OK: true, Detail: "ok", Duration: time.Second}},
		fakeChecker{name: "b", res: CheckResult{Name: "b", Err: boom}},
	}
	results := RunAll(context.Background(), checkers)
	if len(results) != 2 {
		t.Fatalf("RunAll 结果数 = %d, want 2", len(results))
	}
	if !results[0].OK || results[0].Detail != "ok" {
		t.Errorf("results[0] = %#v", results[0])
	}
	if results[1].OK || !errors.Is(results[1].Err, boom) {
		t.Errorf("results[1] = %#v", results[1])
	}
}

func TestFirstNonEmptyLine(t *testing.T) {
	if got := firstNonEmptyLine("\n  hello \nworld\n"); got != "hello" {
		t.Errorf("firstNonEmptyLine = %q", got)
	}
	if got := firstNonEmptyLine("   \n"); got != "" {
		t.Errorf("firstNonEmptyLine(blank) = %q", got)
	}
}

func TestIsLoopbackHost(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"localhost", true},
		{"LOCALHOST", true},
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"127.255.255.254", true},
		{"::1", true},
		{"  localhost  ", true},
		{"example.com", false},
		{"192.168.1.1", false},
		{"10.0.0.5", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isLoopbackHost(c.host); got != c.want {
			t.Errorf("isLoopbackHost(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}

func TestShouldBypassProxy(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		// 回环地址直连。
		{"http://127.0.0.1:11434/", true},
		{"http://localhost:8080/", true},
		{"http://[::1]:9000/", true},
		// 本地服务端口绕过（不区分主机）：Ollama / llama-server / cc-switch。
		{"http://192.168.1.10:11434/", true},
		{"http://example.com:11435/", true},
		{"http://ollama.local:8080/", true},
		{"http://192.168.1.10:15721/", true},
		{"http://cc-switch.local:15721/", true},
		// 其余走代理。
		{"https://api.ipify.org/", false},
		{"https://www.baidu.com/", false},
		{"http://192.168.1.10:9000/", false},
	}
	for _, c := range cases {
		req, err := http.NewRequest(http.MethodGet, c.url, nil)
		if err != nil {
			t.Fatalf("NewRequest(%s): %v", c.url, err)
		}
		if got := shouldBypassProxy(req); got != c.want {
			t.Errorf("shouldBypassProxy(%s) = %v, want %v", c.url, got, c.want)
		}
	}
}

func TestBypassLocalProxy(t *testing.T) {
	proxyURL, err := url.Parse("http://127.0.0.1:7890")
	if err != nil {
		t.Fatal(err)
	}
	proxyFunc := func(*http.Request) (*url.URL, error) { return proxyURL, nil }
	fn := bypassLocalProxy(proxyFunc)

	// 回环地址与本地服务端口应直连（返回 nil 代理）。
	for _, u := range []string{
		"http://127.0.0.1:11434/",
		"http://localhost:8080/",
		"http://[::1]:9000/",
		"http://192.168.1.10:11434/",
		"http://192.168.1.10:15721/",
	} {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			t.Fatalf("NewRequest(%s): %v", u, err)
		}
		got, err := fn(req)
		if err != nil {
			t.Fatalf("bypassLocalProxy(%s) error = %v", u, err)
		}
		if got != nil {
			t.Errorf("bypassLocalProxy(%s) = %v, want nil", u, got)
		}
	}

	// 非回环地址且非本地服务端口应透传代理。
	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org/", nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := fn(req)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.String() != "http://127.0.0.1:7890" {
		t.Errorf("bypassLocalProxy(https://api.ipify.org/) = %v, want proxy", got)
	}
}
