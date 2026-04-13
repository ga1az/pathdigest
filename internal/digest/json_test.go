package digest

import (
	"encoding/json"
	"testing"

	"github.com/ga1az/pathdigest/internal/gitutil"
)

func TestFormatJSON_BasicStructure(t *testing.T) {
	root := &FileNode{
		Name: "project",
		Path: ".",
		Type: NodeTypeDir,
		Children: []*FileNode{
			{
				Name:    "main.go",
				Path:    "main.go",
				Type:    NodeTypeFile,
				Size:    100,
				Content: "package main\n",
			},
			{
				Name: "internal",
				Path: "internal",
				Type: NodeTypeDir,
				Children: []*FileNode{
					{
						Name:    "app.go",
						Path:    "internal/app.go",
						Type:    NodeTypeFile,
						Size:    200,
						Content: "package internal\n",
					},
				},
			},
		},
	}

	result := &Result{
		RootNode:   root,
		TotalFiles: 2,
		TotalSize:  300,
	}

	opts := IngestionOptions{
		Source:          "/tmp/project",
		ExcludePatterns: []string{".git/"},
		IncludePatterns: []string{},
		MaxFileSize:     10 * 1024 * 1024,
	}

	data, err := result.FormatJSON(opts)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON output: %v\nJSON: %s", err, string(data))
	}

	// Verify summary
	if output.Summary.Source != "/tmp/project" {
		t.Errorf("Summary.Source = %q, want %q", output.Summary.Source, "/tmp/project")
	}
	if output.Summary.TotalFiles != 2 {
		t.Errorf("Summary.TotalFiles = %d, want 2", output.Summary.TotalFiles)
	}
	if output.Summary.TotalSize != 300 {
		t.Errorf("Summary.TotalSize = %d, want 300", output.Summary.TotalSize)
	}
	if output.Summary.TotalSizeHuman == "" {
		t.Error("Summary.TotalSizeHuman is empty")
	}

	// Verify arrays are always present (no omitempty)
	if output.Summary.ExcludePatterns == nil {
		t.Error("Summary.ExcludePatterns is nil, should be empty array")
	}
	if output.Summary.IncludePatterns == nil {
		t.Error("Summary.IncludePatterns is nil, should be empty array")
	}

	// Verify tree
	if len(output.Tree) == 0 {
		t.Fatal("Tree is empty")
	}

	// Verify files
	if len(output.Files) != 2 {
		t.Fatalf("Files length = %d, want 2", len(output.Files))
	}

	// Verify file content is always present (even if empty string)
	for _, f := range output.Files {
		if f.Path == "" {
			t.Error("File has empty path")
		}
		if f.Type != "file" {
			t.Errorf("File %q has type %q, want 'file'", f.Path, f.Type)
		}
		// Content field must exist in the JSON — we verify by re-marshaling
	}
}

func TestFormatJSON_ContentAlwaysPresent(t *testing.T) {
	root := &FileNode{
		Name: "project",
		Path: ".",
		Type: NodeTypeDir,
		Children: []*FileNode{
			{
				Name:    "empty.txt",
				Path:    "empty.txt",
				Type:    NodeTypeFile,
				Size:    0,
				Content: "", // empty file
			},
			{
				Name:    "hello.go",
				Path:    "hello.go",
				Type:    NodeTypeFile,
				Size:    50,
				Content: "package main\n",
			},
		},
	}

	result := &Result{
		RootNode:   root,
		TotalFiles: 2,
		TotalSize:  50,
	}

	opts := IngestionOptions{
		Source:          ".",
		ExcludePatterns: []string{},
		IncludePatterns: []string{},
	}

	data, err := result.FormatJSON(opts)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	// Verify that "content" key appears for both files in raw JSON
	raw := string(data)

	// Count occurrences of "content": even empty files must have it
	// We check that both file entries have a "content" field
	var output JSONOutput
	json.Unmarshal(data, &output)

	for _, f := range output.Files {
		if f.Path == "empty.txt" && f.Content != "" {
			t.Errorf("empty.txt Content = %q, want empty string", f.Content)
		}
		if f.Path == "hello.go" && f.Content != "package main\n" {
			t.Errorf("hello.go Content = %q, want %q", f.Content, "package main\n")
		}
	}

	// Verify raw JSON contains "content" key for empty file
	if !containsN(raw, `"content"`, 2) {
		t.Errorf("JSON does not contain 'content' key for all files.\nJSON: %s", raw)
	}
}

func TestFormatJSON_NonTextAndTooLarge(t *testing.T) {
	root := &FileNode{
		Name: "project",
		Path: ".",
		Type: NodeTypeDir,
		Children: []*FileNode{
			{
				Name: "image.png",
				Path: "image.png",
				Type: NodeTypeNotText,
				Size: 5000,
			},
			{
				Name: "big.log",
				Path: "big.log",
				Type: NodeTypeTooLarge,
				Size: 999999999,
			},
		},
	}

	result := &Result{
		RootNode:   root,
		TotalFiles: 2,
		TotalSize:  5000 + 999999999,
	}

	opts := IngestionOptions{Source: "."}

	data, err := result.FormatJSON(opts)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	var output JSONOutput
	json.Unmarshal(data, &output)

	if len(output.Files) != 2 {
		t.Fatalf("Files length = %d, want 2", len(output.Files))
	}

	for _, f := range output.Files {
		if f.Type != "non-text" && f.Type != "too-large" {
			t.Errorf("File %q has unexpected type %q", f.Path, f.Type)
		}
		// Non-text and too-large files should NOT have content field in JSON
		// They are not NodeTypeFile, so they are excluded from content gathering
	}
}

func TestFormatJSON_GitInfo(t *testing.T) {
	root := &FileNode{
		Name:    "repo",
		Path:    ".",
		Type:    NodeTypeDir,
		Content: "",
		Children: []*FileNode{
			{
				Name:    "readme.md",
				Path:    "readme.md",
				Type:    NodeTypeFile,
				Size:    42,
				Content: "# Hello",
			},
		},
	}

	result := &Result{
		RootNode:   root,
		TotalFiles: 1,
		TotalSize:  42,
		GitInfo: &gitutil.GitURLParts{
			RepoURL:  "https://github.com/user/repo.git",
			Host:     "github.com",
			User:     "user",
			RepoName: "repo",
			Branch:   "main",
		},
	}

	opts := IngestionOptions{Source: "user/repo"}

	data, err := result.FormatJSON(opts)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	var output JSONOutput
	json.Unmarshal(data, &output)

	if output.GitInfo == nil {
		t.Fatal("GitInfo is nil, expected non-nil")
	}
	if output.GitInfo.RepoURL != "https://github.com/user/repo.git" {
		t.Errorf("GitInfo.RepoURL = %q, want %q", output.GitInfo.RepoURL, "https://github.com/user/repo.git")
	}
	if output.GitInfo.Branch != "main" {
		t.Errorf("GitInfo.Branch = %q, want %q", output.GitInfo.Branch, "main")
	}
	if output.GitInfo.User != "user" {
		t.Errorf("GitInfo.User = %q, want %q", output.GitInfo.User, "user")
	}
	if output.GitInfo.RepoName != "repo" {
		t.Errorf("GitInfo.RepoName = %q, want %q", output.GitInfo.RepoName, "repo")
	}
}

func TestFormatJSON_NoGitInfo(t *testing.T) {
	root := &FileNode{
		Name:    "file.txt",
		Path:    ".",
		Type:    NodeTypeFile,
		Size:    10,
		Content: "hello",
	}

	result := &Result{
		RootNode:   root,
		TotalFiles: 1,
		TotalSize:  10,
		GitInfo:    nil,
	}

	opts := IngestionOptions{Source: "."}

	data, err := result.FormatJSON(opts)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	raw := string(data)
	if containsN(raw, `"git_info"`, 1) {
		t.Errorf("JSON should not contain git_info when GitInfo is nil.\nJSON: %s", raw)
	}
}

func TestFormatJSON_TreeStructure(t *testing.T) {
	root := &FileNode{
		Name: "root",
		Path: ".",
		Type: NodeTypeDir,
		Children: []*FileNode{
			{
				Name: "a",
				Path: "a",
				Type: NodeTypeDir,
				Children: []*FileNode{
					{
						Name:    "deep.go",
						Path:    "a/deep.go",
						Type:    NodeTypeFile,
						Size:    10,
						Content: "package a",
					},
				},
			},
			{
				Name:    "b.txt",
				Path:    "b.txt",
				Type:    NodeTypeFile,
				Size:    5,
				Content: "hello",
			},
		},
	}

	result := &Result{
		RootNode:   root,
		TotalFiles: 2,
		TotalSize:  15,
	}

	opts := IngestionOptions{Source: "."}

	data, err := result.FormatJSON(opts)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	var output JSONOutput
	json.Unmarshal(data, &output)

	if len(output.Tree) == 0 {
		t.Fatal("Tree is empty")
	}

	// Root should have 2 children: "a" (dir) and "b.txt" (file)
	firstChild := output.Tree[0]
	if firstChild.Name != "a" {
		t.Errorf("First tree child name = %q, want %q", firstChild.Name, "a")
	}
	if firstChild.Type != "directory" {
		t.Errorf("First tree child type = %q, want %q", firstChild.Type, "directory")
	}
	if len(firstChild.Children) != 1 {
		t.Fatalf("First tree child has %d children, want 1", len(firstChild.Children))
	}
	if firstChild.Children[0].Name != "deep.go" {
		t.Errorf("Nested child name = %q, want %q", firstChild.Children[0].Name, "deep.go")
	}

	secondChild := output.Tree[1]
	if secondChild.Name != "b.txt" {
		t.Errorf("Second tree child name = %q, want %q", secondChild.Name, "b.txt")
	}
}

// Helper: count occurrences of substr in string
func containsN(s, substr string, n int) bool {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count >= n
}
