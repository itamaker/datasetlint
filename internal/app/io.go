package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func loadDataset(path string) ([]Example, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open dataset: %w", err)
	}
	defer file.Close()

	var examples []Example
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		if scanner.Text() == "" {
			continue
		}
		var example Example
		if err := json.Unmarshal(scanner.Bytes(), &example); err != nil {
			return nil, fmt.Errorf("decode line %d: %w", line, err)
		}
		examples = append(examples, example)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan dataset: %w", err)
	}
	return examples, nil
}
