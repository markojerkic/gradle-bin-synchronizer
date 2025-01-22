package pkg

import (
	"os"
	"path/filepath"

	"github.com/markojerkic/gradle-bin-synchronizer/pkg/util"
)

func (w *Worker) initialize() {
	err := filepath.Walk(w.SourceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(w.TargetDir, path), 0755)
		} else {
			return util.CopyFile(path, filepath.Join(w.TargetDir, path))
		}
	})

	if err != nil {
		panic(err)
	}

}
