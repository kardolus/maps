package utils

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func FileToBytes(fileName string) ([]byte, error) {
	_, thisFile, _, _ := runtime.Caller(0)

	var (
		urlPath string
		err     error
	)
	if strings.Contains(thisFile, "vendor") {
		urlPath, err = filepath.Abs(path.Join(thisFile, "../../../../../..", "resources", fileName))
	} else {
		urlPath, err = filepath.Abs(path.Join(thisFile, "../..", "resources", fileName))
	}

	if err != nil {
		return nil, err
	}

	return os.ReadFile(urlPath)
}
