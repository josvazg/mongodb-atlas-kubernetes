package akogen

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetFQPath(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to compute absolute path: %w", err)
	}
	dir := filepath.Dir(absPath)
	fullPath := dir
	for {
		modFile := filepath.Join(dir, "go.mod")
		relPath, err := filepath.Rel(dir, fullPath)
		log.Printf("dir=%v", dir)
		log.Printf("rel=%v", relPath)
		if err != nil {
			return "", fmt.Errorf("failed to extract relative module path: %w", err)
		}
		if line, err := readFirstLine(modFile); err == nil {
			if line != "" {
				moduleBase := strings.TrimSpace(strings.TrimPrefix(line, "module "))
				return filepath.Join(moduleBase, relPath), nil
			}
		}
		if dir == "/" || dir == "." {
			break // Reached root directory without finding go.mod
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("no go.mod found")
}

func readFirstLine(filename string) (string, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", fmt.Errorf("failed to read first line: %w", err)
}
