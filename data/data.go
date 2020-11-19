// Package data provides convenience routines to access files in the data
// directory.
package data

import (
	"path/filepath"
	"runtime"
)

// basePath is the root directory of this package.
var basePath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	basePath = filepath.Dir(currentFile)
}

// Path returns the absolute path the given relative file or directory path,
// If rel is already absolute, it is returned unmodified.
func Path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}

	return filepath.Join(basePath, rel)
}
