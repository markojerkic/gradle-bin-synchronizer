package watcher

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/markojerkic/gradle-bin-synchronizer/pkg/util"
)

func (wt *WathingTree) addToWatching(file os.FileInfo) {
	if file.IsDir() {
		wt.watchingDirs[file.Name()] = true
		wt.fileWatcher.Add(file.Name())
		slog.Debug(fmt.Sprintf("Added directory: %s\n", file.Name()))
	} else {
		wt.watchingFiles[file.Name()] = true
		wt.fileWatcher.Add(file.Name())
		slog.Debug(fmt.Sprintf("Added file: %s\n", file.Name()))
	}
}

func (wt *WathingTree) removeFromWatching(file os.FileInfo) {
	if file.IsDir() {
		delete(wt.watchingDirs, file.Name())
		wt.fileWatcher.Remove(file.Name())
		slog.Debug(fmt.Sprintf("Removed directory: %s\n", file.Name()))
	} else {
		delete(wt.watchingFiles, file.Name())
		wt.fileWatcher.Remove(file.Name())
		slog.Debug(fmt.Sprintf("Removed file: %s\n", file.Name()))
	}
}

func (wt *WathingTree) classFileExistsAsJavaFile(path string) bool {
	// Check if the java file exists in src/main/java
	relativePath := strings.TrimPrefix(path, wt.SourceDir)
	javaFilePath := strings.Replace(relativePath, ".class", ".java", 1)

	fullPath := filepath.Join(wt.SourceDir, "../../src/main/java", javaFilePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// Find all mathing files (like App.java and App.class) and if App.java no longer exists, remove App.class
// Do the same for directories and sameextesion files (like application.properties and application.properties)
func (wt *WathingTree) syncWatchAndTargetDir() {
	err := filepath.Walk(wt.SourceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(wt.TargetDir, path), 0755)
		} else if !wt.classFileExistsAsJavaFile(path) && strings.HasSuffix(path, ".class") {
			// Delete the class file if the java file does not exist
			slog.Debug(fmt.Sprintf("Removing file: %s\n", path))
			return os.Remove(path)
		} else {
			return nil
		}
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Error walking the path: %v\n", err))
	}

}

func (wt *WathingTree) watch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {

			case <-ticker.C:
				slog.Debug("Syncing watch and target directories\n")
				wt.syncWatchAndTargetDir()
			}
		}
	}()

	for {
		select {

		case event := <-wt.fileWatcher.Events:
			switch {
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				info, err := os.Stat(event.Name)
				if err != nil {
					slog.Error(fmt.Sprintf("Error getting file info: %v\n", err))
					continue
				}
				wt.removeFromWatching(info)
				// remove the file from the target directory
				err = os.Remove(filepath.Join(wt.TargetDir, info.Name()))
				if err != nil {
					slog.Error(fmt.Sprintf("Error removing file: %v\n", err))
					continue
				}

			case event.Op&fsnotify.Create == fsnotify.Create:
				info, err := os.Stat(event.Name)
				if err != nil {
					slog.Error(fmt.Sprintf("Error getting file info: %v\n", err))
					continue
				}
				wt.addToWatching(info)
				// copy the file to the target directory
				err = os.MkdirAll(filepath.Dir(filepath.Join(wt.TargetDir, info.Name())), 0755)
				if err != nil {
					slog.Error(fmt.Sprintf("Error creating directory: %v\n", err))
					continue
				}
				if !info.IsDir() {
					err = util.CopyFile(info.Name(), filepath.Join(wt.TargetDir, info.Name()))
				}
				if err != nil {
					slog.Error(fmt.Sprintf("Error copying file: %v\n", err))
				}

			case event.Op&fsnotify.Write == fsnotify.Write:
				info, err := os.Stat(event.Name)
				if err != nil {
					slog.Error(fmt.Sprintf("Error getting file info: %v\n", err))
					continue
				}
				slog.Debug(fmt.Sprintf("File %s modified\n", info.Name()))
				// copy the file to the target directory
				err = os.MkdirAll(filepath.Dir(filepath.Join(wt.TargetDir, info.Name())), 0755)
				if err != nil {
					slog.Error(fmt.Sprintf("Error creating directory: %v\n", err))
					continue
				}
				err = util.CopyFile(info.Name(), filepath.Join(wt.TargetDir, info.Name()))
				if err != nil {
					slog.Error(fmt.Sprintf("Error copying file: %v\n", err))
				}
			}

		case err := <-wt.fileWatcher.Errors:
			slog.Error(fmt.Sprintf("Error: %v\n", err))
		}
	}

}
