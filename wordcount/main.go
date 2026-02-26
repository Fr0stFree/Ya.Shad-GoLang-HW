//go:build !solution

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type Counts map[string]int

func (c Counts) Merge(other Counts) {
	for line, amount := range other {
		c[line] += amount
	}
}

func main() {
	paths := os.Args[1:]
	result := make(Counts)

	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		counts, err := countLetters(path)
		if err != nil {
			panic(err)
		}
		result.Merge(counts)
	}

	for line, amount := range result {
		if amount < 2 {
			continue
		}
		fmt.Printf("%d\t%s\n", amount, line)
	}
}

func countLetters(path string) (Counts, error) {
	counts := make(Counts)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		amount, isExists := counts[line]
		if !isExists {
			amount = 0
		}
		counts[line] = amount + 1
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	return counts, nil
}
