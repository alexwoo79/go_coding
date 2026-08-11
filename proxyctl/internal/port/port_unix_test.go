//go:build !windows

package port

import (
	"reflect"
	"testing"
)

func TestParseLsof(t *testing.T) {
	out := "p456\ncclash\nn127.0.0.1:7891\nn[::1]:7891\np123\ncnode\nn127.0.0.1:7890\n"
	got := parseLsof(out)
	want := []Process{
		{PID: 123, Command: "node", Addresses: []string{"127.0.0.1:7890"}},
		{PID: 456, Command: "clash", Addresses: []string{"127.0.0.1:7891", "[::1]:7891"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseLsof = %#v, want %#v", got, want)
	}
}

func TestParseLsofEmpty(t *testing.T) {
	if got := parseLsof(""); len(got) != 0 {
		t.Errorf("parseLsof(\"\") = %#v, want 空", got)
	}
}

func TestParseLsofDuplicatePID(t *testing.T) {
	// 同一 PID 的重复记录只保留一个进程，地址合并去重。
	out := "p7\ncgit\nn127.0.0.1:7\np7\ncgit\nn127.0.0.1:7\n"
	got := parseLsof(out)
	want := []Process{{PID: 7, Command: "git", Addresses: []string{"127.0.0.1:7"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseLsof = %#v, want %#v", got, want)
	}
}

func TestParseLsofBadPID(t *testing.T) {
	// 无法解析的 PID 行应被跳过，不影响后续记录。
	out := "px\ncfoo\nn127.0.0.1:1\np9\ncbar\nn127.0.0.1:9\n"
	got := parseLsof(out)
	want := []Process{{PID: 9, Command: "bar", Addresses: []string{"127.0.0.1:9"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseLsof = %#v, want %#v", got, want)
	}
}
