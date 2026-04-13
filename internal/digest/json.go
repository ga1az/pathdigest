package digest

import (
	"encoding/json"
	"path/filepath"
)

type JSONOutput struct {
	Summary JSONSummary  `json:"summary"`
	Tree    []*JSONNode  `json:"tree"`
	Files   []JSONFile   `json:"files"`
	GitInfo *JSONGitInfo `json:"git_info,omitempty"`
}

type JSONSummary struct {
	Source          string   `json:"source"`
	TotalFiles      int      `json:"total_files"`
	TotalSize       int64    `json:"total_size"`
	TotalSizeHuman  string   `json:"total_size_human"`
	ExcludePatterns []string `json:"exclude_patterns"`
	IncludePatterns []string `json:"include_patterns"`
	MaxFileSize     int64    `json:"max_file_size"`
}

type JSONNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"`
	Size     int64       `json:"size,omitempty"`
	Children []*JSONNode `json:"children,omitempty"`
}

type JSONFile struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type JSONGitInfo struct {
	RepoURL  string `json:"repo_url,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Commit   string `json:"commit,omitempty"`
	User     string `json:"user,omitempty"`
	RepoName string `json:"repo_name,omitempty"`
}

func (r *Result) FormatJSON(opts IngestionOptions) ([]byte, error) {
	output := JSONOutput{
		Summary: JSONSummary{
			Source:          opts.Source,
			TotalFiles:      r.TotalFiles,
			TotalSize:       r.TotalSize,
			TotalSizeHuman:  formatBytes(r.TotalSize),
			ExcludePatterns: opts.ExcludePatterns,
			IncludePatterns: opts.IncludePatterns,
			MaxFileSize:     opts.MaxFileSize,
		},
		Tree:  buildJSONTree(r.RootNode),
		Files: gatherJSONFiles(r.RootNode),
	}

	if r.GitInfo != nil {
		output.GitInfo = &JSONGitInfo{
			RepoURL:  r.GitInfo.RepoURL,
			Branch:   r.GitInfo.Branch,
			Commit:   r.GitInfo.Commit,
			User:     r.GitInfo.User,
			RepoName: r.GitInfo.RepoName,
		}
	}

	return json.MarshalIndent(output, "", "  ")
}

func buildJSONTree(node *FileNode) []*JSONNode {
	if node == nil {
		return nil
	}

	if node.Type == NodeTypeDir && node.Children != nil {
		result := make([]*JSONNode, 0, len(node.Children))
		for _, child := range node.Children {
			result = append(result, fileNodeToJSON(child))
		}
		return result
	}

	return []*JSONNode{fileNodeToJSON(node)}
}

func fileNodeToJSON(node *FileNode) *JSONNode {
	jn := &JSONNode{
		Name: node.Name,
		Path: filepath.ToSlash(node.Path),
		Type: string(node.Type),
		Size: node.Size,
	}

	if node.Type == NodeTypeDir && node.Children != nil {
		jn.Children = make([]*JSONNode, 0, len(node.Children))
		for _, child := range node.Children {
			jn.Children = append(jn.Children, fileNodeToJSON(child))
		}
	}

	return jn
}

func gatherJSONFiles(node *FileNode) []JSONFile {
	var files []JSONFile
	gatherJSONFilesRecursive(node, &files)
	return files
}

func gatherJSONFilesRecursive(node *FileNode, files *[]JSONFile) {
	if node.Type == NodeTypeFile {
		f := JSONFile{
			Path:    filepath.ToSlash(node.Path),
			Size:    node.Size,
			Type:    string(node.Type),
			Content: node.Content, // always include, even if empty
		}
		*files = append(*files, f)
	} else if node.Type == NodeTypeNotText || node.Type == NodeTypeTooLarge {
		*files = append(*files, JSONFile{
			Path: filepath.ToSlash(node.Path),
			Size: node.Size,
			Type: string(node.Type),
		})
	}

	if node.Type == NodeTypeDir {
		for _, child := range node.Children {
			gatherJSONFilesRecursive(child, files)
		}
	}
}
