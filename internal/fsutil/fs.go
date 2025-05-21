package fsutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxBytesToDetectText = 1024
)

func IsTextFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, maxBytesToDetectText)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}
	if n == 0 {
		return true, nil
	}

	content := buffer[:n]

	for _, b := range content {
		if b == 0 {
			return false, nil
		}
	}
	return true, nil
}

func ReadFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetRelativePath(basePath, targetPath string) (string, error) {
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return "", err
	}
	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absTargetPath, absBasePath) {
		return filepath.Base(absTargetPath), nil
	}

	relPath, err := filepath.Rel(absBasePath, absTargetPath)
	if err != nil {
		return filepath.Base(absTargetPath), nil
	}
	return relPath, nil
}
