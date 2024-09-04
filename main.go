package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coalaura/progress"
)

func main() {
	PrintHeader()

	path := strings.Join(os.Args[1:], " ")

	if path == "" {
		fmt.Println("Usage: frm <path>")

		return
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Error: path does not exist.")

		return
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error: failed to get absolute path.")

		return
	}

	fmt.Println(" - Collecting files...")

	files, directories, err := CollectFiles(abs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		return
	}

	workers := calculateMaxWorkers(abs)
	fmt.Printf(" - Using %d workers.\n", workers)

	err = IterateAndDelete("Deleting files", files, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		return
	}

	err = IterateAndDelete("Deleting directories", directories, workers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		return
	}

	fmt.Printf(" - Found %d files and %d directories.\n", len(files), len(directories))

	fmt.Println(" - Done!")
}

func IterateAndDelete(label string, files []string, workers int) error {
	queue := NewQueue(workers)
	bar := progress.NewProgressBarWithTheme(label, len(files), progress.ThemeGradientUnicode)

	defer queue.Stop()
	defer bar.Abort()

	for _, file := range files {
		err := queue.Work(func(p string) error {
			if err := os.Remove(p); err != nil {
				return err
			}

			bar.Increment()

			return nil
		}, file)

		if err != nil {
			return err
		}
	}

	err := queue.Stop()
	bar.Stop()

	return err
}

func CollectFiles(path string) ([]string, []string, error) {
	var (
		files []string
		dirs  []string
	)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Add root dir itself
	dirs = append(dirs, path)

	sort.Slice(files, func(i, j int) bool {
		return len(files[i]) > len(files[j])
	})

	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	return files, dirs, nil
}
