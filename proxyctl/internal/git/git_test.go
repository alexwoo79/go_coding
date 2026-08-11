package git

import (
	"errors"
	"reflect"
	"testing"
)

// fakeRunner 记录调用并返回预设结果。
type fakeRunner struct {
	run   func(args ...string) ([]byte, error)
	calls [][]string
}

func (f *fakeRunner) Run(args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	if f.run == nil {
		return nil, nil
	}
	return f.run(args...)
}

func TestGetValue(t *testing.T) {
	r := &fakeRunner{run: func(args ...string) ([]byte, error) {
		return []byte("http://127.0.0.1:7890\n"), nil
	}}
	c := NewWithRunner(r)

	got, err := c.Get("http.proxy")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "http://127.0.0.1:7890" {
		t.Errorf("Get = %q", got)
	}
	want := []string{"config", "--global", "--get", "http.proxy"}
	if !reflect.DeepEqual(r.calls[0], want) {
		t.Errorf("调用参数 = %v, want %v", r.calls[0], want)
	}
}

func TestGetNotSet(t *testing.T) {
	c := NewWithRunner(&fakeRunner{run: func(args ...string) ([]byte, error) {
		return nil, ErrNotSet
	}})

	got, err := c.Get("http.proxy")
	if err != nil {
		t.Fatalf("Get 未配置项: %v", err)
	}
	if got != "" {
		t.Errorf("Get 未配置项 = %q, want 空串", got)
	}
}

func TestGetError(t *testing.T) {
	c := NewWithRunner(&fakeRunner{run: func(args ...string) ([]byte, error) {
		return nil, errors.New("git 不存在")
	}})

	if _, err := c.Get("http.proxy"); err == nil {
		t.Fatal("Get 应返回错误")
	}
}

func TestSet(t *testing.T) {
	r := &fakeRunner{}
	if err := NewWithRunner(r).Set("http.proxy", "http://127.0.0.1:7890"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	want := []string{"config", "--global", "http.proxy", "http://127.0.0.1:7890"}
	if !reflect.DeepEqual(r.calls[0], want) {
		t.Errorf("调用参数 = %v, want %v", r.calls[0], want)
	}
}

func TestUnsetNotSetIsOK(t *testing.T) {
	c := NewWithRunner(&fakeRunner{run: func(args ...string) ([]byte, error) {
		return nil, ErrNotSet
	}})
	if err := c.Unset("http.proxy"); err != nil {
		t.Fatalf("Unset 未配置项应成功: %v", err)
	}
}

func TestUnsetError(t *testing.T) {
	c := NewWithRunner(&fakeRunner{run: func(args ...string) ([]byte, error) {
		return nil, errors.New("boom")
	}})
	if err := c.Unset("http.proxy"); err == nil {
		t.Fatal("Unset 应返回错误")
	}
}

func TestSnapshot(t *testing.T) {
	values := map[string]string{"http.proxy": "http://127.0.0.1:7890"}
	c := NewWithRunner(&fakeRunner{run: func(args ...string) ([]byte, error) {
		if args[3] == "http.proxy" {
			return []byte(values["http.proxy"] + "\n"), nil
		}
		return nil, ErrNotSet
	}})

	snap, err := c.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if len(snap.Values) != 1 || snap.Values["http.proxy"] != "http://127.0.0.1:7890" {
		t.Errorf("Snapshot.Values = %#v", snap.Values)
	}
}

func TestRestore(t *testing.T) {
	r := &fakeRunner{}
	snap := Snapshot{Values: map[string]string{
		"http.proxy": "http://127.0.0.1:7890",
	}}
	if err := NewWithRunner(r).Restore(snap); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(r.calls) != 2 {
		t.Fatalf("Restore 应执行 2 次调用，实际 %d", len(r.calls))
	}
	if !reflect.DeepEqual(r.calls[0], []string{"config", "--global", "http.proxy", "http://127.0.0.1:7890"}) {
		t.Errorf("calls[0] = %v", r.calls[0])
	}
	if !reflect.DeepEqual(r.calls[1], []string{"config", "--global", "--unset", "https.proxy"}) {
		t.Errorf("calls[1] = %v", r.calls[1])
	}
}
