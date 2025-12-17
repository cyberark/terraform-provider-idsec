// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dirs := []string{
		"../docs/resources",
		"../docs/data-sources",
	}
	dir, _ := os.Getwd()
	fmt.Printf("%s\n", dir)
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("Skipping missing directory: %s\n", dir)
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}

			filename := filepath.Base(path)
			name := strings.TrimSuffix(filename, ".md")

			// must contain prefix_
			parts := strings.SplitN(name, "_", 2)
			if len(parts) < 2 {
				fmt.Printf("Skipping (no prefix): %s\n", path)
				return nil
			}

			prefix := strings.ToLower(parts[0])
			targetDir := filepath.Join(dir, prefix)
			err = os.MkdirAll(targetDir, 0o700)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}

			targetPath := filepath.Join(targetDir, filename)

			fmt.Printf("Moving %s → %s\n", path, targetPath)

			// Move the file
			if err := os.Rename(path, targetPath); err != nil {
				return fmt.Errorf("failed to move %s → %s: %w", path, targetPath, err)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error processing %s: %v\n", dir, err)
		}
	}

	fmt.Println("Document reorganization completed.")
}
