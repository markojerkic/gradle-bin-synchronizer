package watcher

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type WathingTree struct {
	SourceDir string
	TargetDir string

	watchingFiles map[string]bool
	watchingDirs  map[string]bool

	fileWatcher *fsnotify.Watcher
}

func (wt *WathingTree) initialize() {
	err := filepath.Walk(wt.SourceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			wt.watchingDirs[path] = true
			wt.fileWatcher.Add(path)
			slog.Debug(fmt.Sprintf("Added directory: %s\n", path))
		} else {
			wt.watchingFiles[path] = true
			wt.fileWatcher.Add(path)
			slog.Debug(fmt.Sprintf("Added file: %s\n", path))
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}

func NewWatchingTree(sourceDir, targetDir string) *WathingTree {
	fsWatcher, err := fsnotify.NewWatcher()

	if err != nil {
		panic(err)
	}

	tree := &WathingTree{
		SourceDir:     sourceDir,
		TargetDir:     targetDir,
		watchingFiles: make(map[string]bool),
		watchingDirs:  make(map[string]bool),
		fileWatcher:   fsWatcher,
	}
	tree.initialize()
	tree.watch()

	return tree
}
