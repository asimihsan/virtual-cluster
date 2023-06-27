package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func StatUpward(relPath string, verbose bool) (os.FileInfo, string, error) {
	if filepath.IsAbs(relPath) {
		stat, err := os.Stat(relPath)
		return stat, relPath, err
	}

	absPath, err := filepath.Abs(".")
	if err != nil {
		return nil, "", err
	}

	var errorsEncountered []string

	for {
		testPath := filepath.Join(absPath, relPath)
		stat, err := os.Stat(testPath)
		if err == nil {
			return stat, testPath, nil
		} else {
			if verbose {
				errorsEncountered = append(errorsEncountered, fmt.Sprintf("Error at %s: %v", testPath, err))
			}
		}

		if absPath == filepath.Dir(absPath) {
			break
		}

		absPath = filepath.Dir(absPath)
	}

	if verbose {
		return nil, "", fmt.Errorf("file or directory not found in any parent directory; errors encountered:\n%s", strings.Join(errorsEncountered, "\n"))
	}

	return nil, "", errors.New("file or directory not found in any parent directory")
}

func ReadFileUpward(relPath string, verbose bool) ([]byte, error) {
	stat, foundPath, err := StatUpward(relPath, verbose)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to stat path: %s", relPath)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", foundPath)
	}

	return os.ReadFile(foundPath)
}
