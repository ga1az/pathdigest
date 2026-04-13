package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRelativePath(t *testing.T) {
	// Create temp dirs for realistic path testing
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "project")
	nestedDir := filepath.Join(baseDir, "src", "lib")
	similarDir := filepath.Join(tmpDir, "project2", "file.txt")

	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested test directory %q: %v", nestedDir, err)
	}
	if err := os.MkdirAll(filepath.Dir(similarDir), 0755); err != nil {
		t.Fatalf("failed to create similar test directory %q: %v", filepath.Dir(similarDir), err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create nested test file %q: %v", filepath.Join(nestedDir, "main.go"), err)
	}
	if err := os.WriteFile(similarDir, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create similar test file %q: %v", similarDir, err)
	}

	tests := []struct {
		name     string
		base     string
		target   string
		expected string
	}{
		{
			name:     "nested file",
			base:     baseDir,
			target:   filepath.Join(nestedDir, "main.go"),
			expected: filepath.Join("src", "lib", "main.go"),
		},
		{
			name:     "same directory",
			base:     baseDir,
			target:   baseDir,
			expected: ".",
		},
		{
			name:     "similar prefix should not match",
			base:     baseDir,
			target:   similarDir,
			expected: "file.txt", // Should fall back to basename, not "2/file.txt"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRelativePath(tt.base, tt.target)
			if err != nil {
				t.Fatalf("GetRelativePath(%q, %q) returned error: %v", tt.base, tt.target, err)
			}
			if got != tt.expected {
				t.Errorf("GetRelativePath(%q, %q) = %q, want %q", tt.base, tt.target, got, tt.expected)
			}
		})
	}
}
