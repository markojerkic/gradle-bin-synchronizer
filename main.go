package main

import (
	"flag"
	"github.com/markojerkic/gradle-bin-synchronizer/pkg/watcher"
	"log/slog"
)

func main() {
	// These are the source and target directories we'll be watching and syncing
	pSourceDir := flag.String("watchDir", "bin/main", "Directory to watch for changes")
	pTargetDir := flag.String("syncDir", "build/classes/java/main", "Directory to sync changes to")
	pIsDebug := flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	sourceDir := *pSourceDir
	targetDir := *pTargetDir
	isDebug := *pIsDebug

	if isDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	watcher.NewWatchingTree(sourceDir, targetDir)
}
