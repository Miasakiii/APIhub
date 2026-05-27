package jsonl

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileMeta describes a single JSONL file in the projects directory.
type FileMeta struct {
	Path         string
	SizeBytes    int64
	ModifiedAt   int64 // unix timestamp, used for incremental sync
}

// ScanProjects walks ~/.claude/projects/ and finds all .jsonl files.
// Returns them sorted by modification time (newest first).
func ScanProjects(baseDir string) ([]FileMeta, error) {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".claude", "projects")
	}

	var files []FileMeta
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			files = append(files, FileMeta{
				Path:       path,
				SizeBytes:  info.Size(),
				ModifiedAt: info.ModTime().Unix(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt > files[j].ModifiedAt
	})
	return files, nil
}

// OpenFile returns a *os.File for reading the JSONL file.
func OpenFile(path string) (*os.File, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

// OpenFileAt returns a file seeked to the given byte offset (incremental sync).
func OpenFileAt(path string, offset int64) (*os.File, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	if _, err := f.Seek(offset, 0); err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}
