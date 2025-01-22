package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// These are the source and target directories we'll be watching and syncing
	sourceDir := *flag.String("watchDir", "bin/main", "Directory to watch for changes")
	targetDir := *flag.String("syncDir", "build/classes/java/main", "Directory to sync changes to")
	isDebug := *flag.Bool("debug", false, "Debug mode")
	flag.Parse()
	if isDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

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

	slog.Info(fmt.Sprintf("Watching %s for changes...\n", sourceDir))

	// Start watching for changes
	for {
		select {
		case event := <-watcher.Events:
			// Get the relative path to maintain directory structure
			relPath, err := filepath.Rel(sourceDir, event.Name)
			if err != nil {
				slog.Error(fmt.Sprintf("Error getting relative path: %v\n", err))
				continue
			}
			targetPath := filepath.Join(targetDir, relPath)

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write:
				// File was modified
				err = copyFile(event.Name, targetPath)
				if err != nil {
					slog.Error(fmt.Sprintf("Error copying modified file: %v\n", err))
				} else {
					slog.Debug(fmt.Sprintf("Copied modified file: %s\n", relPath))
				}

			case event.Op&fsnotify.Create == fsnotify.Create:
				// Something was created
				info, err := os.Stat(event.Name)
				if err != nil {
					slog.Error(fmt.Sprintf("Error getting file info: %v\n", err))
					continue
				}

				if info.IsDir() {
					// Add all files recursively to the watcher
					addAllFilesToWatcher(watcher, event.Name)
					os.MkdirAll(targetPath, 0755)
					slog.Debug(fmt.Sprintf("Created directory: %s\n", relPath))
				} else {
					// New file created
					err = copyFile(event.Name, targetPath)
					if err != nil {
						slog.Error(fmt.Sprintf("Error copying new file: %v\n", err))
					} else {
						slog.Info(fmt.Sprintf("Copied new file: %s\n", relPath))
					}
				}

			case event.Op&fsnotify.Remove == fsnotify.Remove:
				// Something was removed
				err = os.RemoveAll(targetPath)
				if err != nil {
					slog.Error(fmt.Sprintf("Error removing: %v\n", err))
				} else {
					slog.Info(fmt.Sprintf("Removed: %s\n", relPath))
					// Remove the directory from the watcher
					if err = watcher.Remove(event.Name); err != nil {
						slog.Error(fmt.Sprintf("Error deleting %+v", err))
					}
				}
			}

		case err := <-watcher.Errors:
			slog.Error(fmt.Sprintf("Watcher error: %v\n", err))
		}
	}
}

func addAllFilesToWatcher(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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
