package main_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	binName  = "todo"
	fileName = ".todo.json"
)

func TestMain(m *testing.M) {
	fmt.Println("Building tool ...")
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	build := exec.Command("go", "build", "-o", binName)
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot build tool %s: %s", binName, err)
		os.Exit(1)
	}
	fmt.Println("Running tests ...")
	result := m.Run()
	fmt.Println("Cleaning up ...")
	os.Remove(binName)
	os.Remove(fileName)
	os.Exit(result)
}
func TestToDoCLI(t *testing.T) {
	task := "Test task Number 1"
	_ = os.Remove(fileName)
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmdPath := filepath.Join(dir, binName)
	t.Run("AddNewTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", task)
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})
	task2 := "Test task Number 2"
	t.Run("AddNewTaskFromSTDIN", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, task2)
		}()
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ListTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf(" 1: %s\n 2: %s\n", task, task2)
		if expected != string(out) {
			t.Errorf("Expected %q, got %q", expected, string(out))
		}

	})

	t.Run("CompleteTaskShowsAsDone", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-complete", "1")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("x 1: %s\n 2: %s\n", task, task2)
		if expected != string(out) {
			t.Errorf("Expected %q, got %q", expected, string(out))
		}
	})
}
