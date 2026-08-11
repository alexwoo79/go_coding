package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	//Defining a boolean flag -l to count lines instead of words. The flag is set to false by default, meaning that the program will count words unless the -l flag is provided when running the program.
	lines := flag.Bool("l", false, "Count lines")
	//Parse the command-line flags provided by the user. This step is necessary to process any flags (like -l) that were specified when running the program.
	flag.Parse()
	//Calling the count function and passing os.Stdin as an argument to read from standard input
	//received from the standard input and printing the word count to the console
	count, err := count(os.Stdin, *lines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(count)
}

func count(r io.Reader, countLines bool) (int, error) {
	scanner := bufio.NewScanner(r)
	if !countLines {
		scanner.Split(bufio.ScanWords)
	}
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}
