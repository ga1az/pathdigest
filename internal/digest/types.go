package digest

import (
	"io/fs"

	"github.com/ga1az/pathdigest/internal/gitutil"
)

type IngestionOptions struct {
	Source          string
	OutputFile      string
	MaxFileSize     int64
	ExcludePatterns []string
	IncludePatterns []string
	Branch          string
}

type FileNodeType string

const (
	NodeTypeFile     FileNodeType = "file"
	NodeTypeDir      FileNodeType = "directory"
	NodeTypeSymlink  FileNodeType = "symlink"
	NodeTypeNotText  FileNodeType = "non-text"
	NodeTypeTooLarge FileNodeType = "too-large"
	NodeTypeExcluded FileNodeType = "excluded"
)

type FileNode struct {
	Name     string
	Path     string
	FullPath string
	Type     FileNodeType
	Size     int64
	Mode     fs.FileMode
	Content  string
	Children []*FileNode
	Error    error
	Depth    int
}

type Result struct {
	Summary       string
	TreeStructure string
	FileContents  string
	RootNode      *FileNode
	TotalFiles    int
	TotalSize     int64
	TokenCount    int
	GitInfo       *gitutil.GitURLParts
}
