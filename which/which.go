package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: which <command>")
		os.Exit(1)
	}

	name := os.Args[1]

	// If the argument contains a path separator, check it directly
	// instead of searching PATH.
	if strings.ContainsRune(name, filepath.Separator) || strings.ContainsRune(name, '/') {
		if !isExecutable(name) {
			os.Exit(1)
		}
		fmt.Println(name)
		return
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		for _, candidate := range executableNames(name) {
			fullPath := filepath.Join(dir, candidate)
			if isExecutable(fullPath) {
				fmt.Println(fullPath)
				return
			}
		}
	}
	os.Exit(1)
}

// executableNames returns the file names to try for the given command.
// On Windows, "go" must be tried as "go.EXE", "go.BAT", etc. per PATHEXT.
func executableNames(name string) []string {
	if runtime.GOOS != "windows" {
		return []string{name}
	}
	if filepath.Ext(name) != "" {
		return []string{name}
	}
	var names []string
	for _, ext := range strings.Split(os.Getenv("PATHEXT"), ";") {
		if ext != "" {
			names = append(names, name+ext)
		}
	}
	return append(names, name)
}

// isExecutable reports whether path is a regular file that can be run.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	if runtime.GOOS != "windows" {
		// Unix-like systems mark executables with the execute permission bits.
		return info.Mode()&0111 != 0
	}
	// On Windows the permission bits are not set for executables; a file is
	// runnable only when its extension is listed in PATHEXT.
	ext := strings.ToLower(filepath.Ext(path))
	for _, ext2 := range strings.Split(os.Getenv("PATHEXT"), ";") {
		if ext == strings.ToLower(ext2) {
			return true
		}
	}
	return false
}
