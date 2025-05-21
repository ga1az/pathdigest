package digest

import (
	"testing"
)

func TestIsPathMatchWithInfo(t *testing.T) {
	tests := []struct {
		name         string
		relativePath string
		isDir        bool
		patterns     []string
		wantMatch    bool
	}{
		// Cases with file patterns
		{"Exact file", "file.txt", false, []string{"file.txt"}, true},
		{"Exact file no match", "other.txt", false, []string{"file.txt"}, false},
		{"Wildcard *.txt", "doc.txt", false, []string{"*.txt"}, true},
		{"Wildcard *.txt no match extension", "doc.md", false, []string{"*.txt"}, false},
		{"Wildcard path complete", "src/file.go", false, []string{"src/*.go"}, true},
		{"Wildcard path complete no match dir", "lib/file.go", false, []string{"src/*.go"}, false},
		{"Wildcard only filename", "main.go", false, []string{"main.*"}, true},
		{"Multiple patterns, one matches", "image.jpg", false, []string{"*.png", "*.jpg", "*.gif"}, true},
		{"Multiple patterns, none match", "image.bmp", false, []string{"*.png", "*.jpg", "*.gif"}, false},
		{"Pattern only filename, file in subdir", "src/config.json", false, []string{"config.json"}, true},

		// Cases with directory patterns (pattern ends in /)
		{"Exact dir", "src", true, []string{"src/"}, true},
		{"Exact dir with /", "src/", true, []string{"src/"}, true},
		{"Exact dir no match", "docs", true, []string{"src/"}, false},
		{"File inside dir pattern", "node_modules/package/file.js", false, []string{"node_modules/"}, true},
		{"Dir inside dir pattern", "vendor/lib/sublib", true, []string{"vendor/"}, true},
		{"Dir pattern no match with file", "main.go", false, []string{"src/"}, false},
		{"Deep dir pattern", "a/b/c/d.txt", false, []string{"a/b/"}, true},
		{"Dir pattern not so deep", "a/other.txt", false, []string{"a/b/"}, false},
		{"Dir pattern in root", "build", true, []string{"build/"}, true},
		{"File in dir pattern in root", "dist/bundle.js", false, []string{"dist/"}, true},

		// Cases with relative path starting with ./
		{"./file.txt with *.txt", "./file.txt", false, []string{"*.txt"}, true},
		{"./src/file.go with src/*.go", "./src/file.go", false, []string{"src/*.go"}, true},
		{"./src/ with src/", "./src/", true, []string{"src/"}, true},

		// Cases with more complex patterns or conflicts (for thinking)
		{"Pattern could be file or dir", "config", false, []string{"config"}, true},
		{"Pattern could be file or dir", "config", true, []string{"config"}, true},
		{"Pattern of dir vs file with same name", "data", false, []string{"data/"}, false},
		{"Pattern of file vs dir with same name", "data/", true, []string{"data"}, true},

		// Empty cases
		{"No patterns", "file.txt", false, []string{}, false},
		{"Empty path, no patterns", "", false, []string{}, false},
		{"Empty path, with pattern", "", false, []string{"*.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotMatch := isPathMatchWithInfo(tt.relativePath, tt.isDir, tt.patterns)
			if gotMatch != tt.wantMatch {
				t.Errorf("isPathMatchWithInfo(%q, %t, %v) = %v, want %v", tt.relativePath, tt.isDir, tt.patterns, gotMatch, tt.wantMatch)
			}
		})
	}
}
