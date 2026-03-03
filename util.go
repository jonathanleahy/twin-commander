package main

import (
	"io"
	"os"
	"unicode/utf8"
)

// readFileHead reads up to maxBytes from the beginning of a file.
func readFileHead(path string, maxBytes int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxBytes))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// isBinary returns true if the data appears to be binary (contains null bytes
// or is not valid UTF-8 after checking the first chunk).
func isBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check first 512 bytes for null bytes
	check := data
	if len(check) > 512 {
		check = check[:512]
	}

	for _, b := range check {
		if b == 0 {
			return true
		}
	}

	// Check if it's valid UTF-8
	if !utf8.Valid(check) {
		return true
	}

	return false
}
