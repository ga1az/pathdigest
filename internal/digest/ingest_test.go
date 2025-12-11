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

		{"./file.txt with *.txt", "./file.txt", false, []string{"*.txt"}, true},
		{"./src/file.go with src/*.go", "./src/file.go", false, []string{"src/*.go"}, true},
		{"./src/ with src/", "./src/", true, []string{"src/"}, true},

		{"Pattern could be file or dir", "config", false, []string{"config"}, true},
		{"Pattern could be file or dir", "config", true, []string{"config"}, true},
		{"Pattern of dir vs file with same name", "data", false, []string{"data/"}, false},
		{"Pattern of file vs dir with same name", "data/", true, []string{"data"}, true},

		{"No patterns", "file.txt", false, []string{}, false},
		{"Empty path, no patterns", "", false, []string{}, false},
		{"Empty path, with pattern", "", false, []string{"*.txt"}, false},

		{"Dir ancestor of nested pattern", "docs", true, []string{"docs/src/"}, true},
		{"Dir ancestor of deeply nested pattern", "a", true, []string{"a/b/c/"}, true},
		{"Dir match middle of nested pattern", "a/b", true, []string{"a/b/c/"}, true},
		{"Dir NOT ancestor of other nested pattern", "lib", true, []string{"docs/src/"}, false},
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

func TestShouldProcessDirForInclude(t *testing.T) {
	tests := []struct {
		name         string
		relativePath string
		patterns     []string
		wantMatch    bool
	}{
		{"Dir matches pattern exactly", "src", []string{"src/"}, true},
		{"Dir is descendant of pattern", "src/components", []string{"src/"}, true},
		{"Dir is ancestor of pattern", "docs", []string{"docs/src/"}, true},
		{"Dir is ancestor of deeply nested pattern", "a", []string{"a/b/c/d/"}, true},
		{"Dir in middle of pattern path", "a/b", []string{"a/b/c/d/"}, true},
		{"Dir does not match pattern", "lib", []string{"src/"}, false},
		{"Dir does not match nested pattern", "lib", []string{"docs/src/"}, false},

		{"Dir with file pattern *.go", "src", []string{"*.go"}, true},
		{"Dir with file pattern *.md", "docs", []string{"*.md"}, true},
		{"Nested dir with file pattern", "a/b/c", []string{"*.txt"}, true},
		{"Any dir with any file pattern", "anything", []string{"*.log"}, true},

		{"Dir matches one of multiple patterns", "src", []string{"docs/", "src/"}, true},
		{"Dir ancestor of one of multiple patterns", "docs", []string{"docs/api/", "lib/"}, true},
		{"Dir with mixed file and dir patterns", "lib", []string{"src/", "*.go"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch := shouldProcessDirForInclude(tt.relativePath, tt.patterns)
			if gotMatch != tt.wantMatch {
				t.Errorf("shouldProcessDirForInclude(%q, %v) = %v, want %v", tt.relativePath, tt.patterns, gotMatch, tt.wantMatch)
			}
		})
	}
}
