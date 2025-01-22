package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// These are the source and target directories we'll be watching and syncing
	sourceDir := "bin/main"
	targetDir := "build/classes/java/main"

	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// First, let's do an initial sync of the directories
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path to maintain directory structure
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relPath)

		if info.IsDir() {
			// Create directory in target
			return os.MkdirAll(targetPath, 0755)
		}

		// Copy the file
		return copyFile(path, targetPath)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add all subdirectories to the watcher
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Watching %s for changes...\n", sourceDir)

	// Start watching for changes
	for {
		select {
		case event := <-watcher.Events:
			// Get the relative path to maintain directory structure
			relPath, err := filepath.Rel(sourceDir, event.Name)
			if err != nil {
				log.Printf("Error getting relative path: %v\n", err)
				continue
			}
			targetPath := filepath.Join(targetDir, relPath)

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write:
				// File was modified
				err = copyFile(event.Name, targetPath)
				if err != nil {
					log.Printf("Error copying modified file: %v\n", err)
				} else {
					fmt.Printf("Copied modified file: %s\n", relPath)
				}

			case event.Op&fsnotify.Create == fsnotify.Create:
				// Something was created
				info, err := os.Stat(event.Name)
				if err != nil {
					log.Printf("Error getting file info: %v\n", err)
					continue
				}

				if info.IsDir() {
					// Add all files recursively to the watcher
					addAllFilesToWatcher(watcher, event.Name)
					os.MkdirAll(targetPath, 0755)
					fmt.Printf("Created directory: %s\n", relPath)
				} else {
					// New file created
					err = copyFile(event.Name, targetPath)
					if err != nil {
						log.Printf("Error copying new file: %v\n", err)
					} else {
						fmt.Printf("Copied new file: %s\n", relPath)
					}
				}

			case event.Op&fsnotify.Remove == fsnotify.Remove:
				// Something was removed
				err = os.RemoveAll(targetPath)
				if err != nil {
					log.Printf("Error removing: %v\n", err)
				} else {
					fmt.Printf("Removed: %s\n", relPath)
					// Remove the directory from the watcher
					if err = watcher.Remove(event.Name); err != nil {
						log.Printf("Error deleting %+v", err)
					}
				}
			}

		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v\n", err)
		}
	}
}

func addAllFilesToWatcher(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		log.Print("File %+v\n", info)
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}

// copyFile handles the file copying operation, creating parent directories if needed
func copyFile(src, dst string) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create destination file
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// Copy the contents
	_, err = io.Copy(destination, source)
	return err
}
