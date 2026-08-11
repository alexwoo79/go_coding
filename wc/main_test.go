package main

import (
	"bytes"
	"testing"
)

func TestCountLines(t *testing.T) {
	// Define test cases with input strings and expected word counts
	b := bytes.NewBufferString("word1 word2 word3 word4\nline2\nline3 word1")
	exp := 3
	// Call the count function with the test input and check the result lines
	res, err := count(b, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res != exp {
		t.Errorf("Expected %d lines, but got %d", exp, res)
	}
}

func TestCountWords(t *testing.T) {
	// Define test cases with input strings and expected word counts
	b := bytes.NewBufferString("word1 word2 word3 word4\nline2\nline3 word1")
	exp := 7
	// Call the count function with the buffer and false to count words
	res, err := count(b, false)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res != exp {
		t.Errorf("Expected %d words, but got %d", exp, res)
	}
}
