package internal

import (
	"os"
	"strings"
)

type FileStorage struct {
	FilePath string
}

func (fs *FileStorage) ReadPreviousResults() ([]string, error) {
	readBytes, err := os.ReadFile(fs.FilePath)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(readBytes), "\n"), nil
}

func (fs *FileStorage) WriteResults(results []string) error {
	return os.WriteFile(fs.FilePath, []byte(strings.Join(results, "\n")), 0644)
}
