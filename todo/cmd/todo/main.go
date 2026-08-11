package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/alexwoo79/go_coding/todo"
)

// Hardcoding the file name
var todoFileName = ".todo.json"

func init() {
	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}
}

func normalizeLegacyFlags(args []string) []string {
	normalized := make([]string, len(args))
	copy(normalized, args)

	for i, arg := range normalized {
		switch arg {
		case "-add", "-task", "-list", "-complete":
			normalized[i] = "-" + arg
		}
	}

	return normalized
}

func getTask(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	s := bufio.NewScanner(r)
	s.Scan()
	if err := s.Err(); err != nil {
		return "", err
	}
	if len(s.Text()) == 0 {
		return "", fmt.Errorf("no task provided")
	}
	return s.Text(), nil
}

func main() {
	//parse the command line arguments
	add := pflag.Bool("add", false, "Add a task to the to-do list")
	task := pflag.String("task", "", "Task to be added to the to-do list")
	list := pflag.Bool("list", false, "List all tasks")
	complete := pflag.Int("complete", 0, "Item to be completed")
	os.Args = normalizeLegacyFlags(os.Args)
	pflag.Parse()
	l := &todo.List{}

	//Use the get method to read to-do items from file
	if err := l.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Decide what to do based on the number of arguments provided
	//for no extra arguments,print the list
	switch {
	case *add:
		t, err := getTask(os.Stdin, pflag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		l.Add(t)
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *list:

		fmt.Print(l)
		//concatenate all provided arguments with a space and
		//add to the list as an item
	case *complete > 0:
		if err := l.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *task != "":
		l.Add(*task)
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid option")
		os.Exit(1)
	}
}
