package digest

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	fileSeparator = "================================================\n"
)

func (r *Result) FormatOutput(opts IngestionOptions) {
	var sbSummary, sbTree, sbContent strings.Builder

	sbSummary.WriteString(createSummaryPrefix(opts, r.RootNode.Type == NodeTypeFile))
	if r.RootNode.Type == NodeTypeDir {
		sbSummary.WriteString(fmt.Sprintf("Files analyzed: %d\n", r.TotalFiles))
		sbSummary.WriteString(fmt.Sprintf("Total size: %s\n", formatBytes(r.TotalSize)))
	} else if r.RootNode.Type == NodeTypeFile {
		sbSummary.WriteString(fmt.Sprintf("File: %s\n", r.RootNode.Name))
		sbSummary.WriteString(fmt.Sprintf("Size: %s\n", formatBytes(r.RootNode.Size)))
		if r.RootNode.Type == NodeTypeFile {
			sbSummary.WriteString(fmt.Sprintf("Lines: %d\n", strings.Count(r.RootNode.Content, "\n")+1))
		}
	}
	r.Summary = sbSummary.String()

	if r.RootNode.Type == NodeTypeDir {
		sbTree.WriteString("Directory structure:\n")
		buildTreeStructure(&sbTree, r.RootNode, "", true)
	} else if r.RootNode.Type == NodeTypeFile {
		sbTree.WriteString("File processed:\n")
		sbTree.WriteString(fmt.Sprintf("└── %s\n", r.RootNode.Name))
	}
	r.TreeStructure = sbTree.String()

	gatherFileContents(&sbContent, r.RootNode)
	r.FileContents = sbContent.String()
}

func createSummaryPrefix(opts IngestionOptions, isSingleFile bool) string {
	var parts []string
	if isSingleFile {
		parts = append(parts, fmt.Sprintf("Source File: %s", opts.Source))
	} else {
		parts = append(parts, fmt.Sprintf("Source Directory: %s", opts.Source))
	}
	if len(opts.IncludePatterns) > 0 {
		parts = append(parts, fmt.Sprintf("Include Patterns: %s", strings.Join(opts.IncludePatterns, ", ")))
	}
	if len(opts.ExcludePatterns) > 0 {
		parts = append(parts, fmt.Sprintf("Exclude Patterns: %s", strings.Join(opts.ExcludePatterns, ", ")))
	}
	if opts.MaxFileSize > 0 {
		parts = append(parts, fmt.Sprintf("Max File Size: %s", formatBytes(opts.MaxFileSize)))
	}

	return strings.Join(parts, "\n") + "\n"
}

func buildTreeStructure(sb *strings.Builder, node *FileNode, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	displayName := node.Name
	switch node.Type {
	case NodeTypeDir:
		displayName += "/"
	case NodeTypeSymlink:
		displayName += " (symlink)"
	case NodeTypeNotText:
		displayName += " (non-text)"
	case NodeTypeTooLarge:
		displayName += fmt.Sprintf(" (too large: %s)", formatBytes(node.Size))
	case NodeTypeExcluded:
		displayName += " (excluded)"
	}

	sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, displayName))

	if node.Type == NodeTypeDir && len(node.Children) > 0 {
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		for i, child := range node.Children {
			buildTreeStructure(sb, child, newPrefix, i == len(node.Children)-1)
		}
	}
}

func gatherFileContents(sb *strings.Builder, node *FileNode) {
	if node.Type == NodeTypeFile && node.Content != "" {
		sb.WriteString(fileSeparator)
		sb.WriteString(fmt.Sprintf("File: %s\n", filepath.ToSlash(node.Path)))
		sb.WriteString(fileSeparator)
		sb.WriteString(node.Content)
		if !strings.HasSuffix(node.Content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	} else if node.Type == NodeTypeNotText || node.Type == NodeTypeTooLarge {
		sb.WriteString(fileSeparator)
		sb.WriteString(fmt.Sprintf("File: %s (%s - content not included)\n", filepath.ToSlash(node.Path), node.Type))
		sb.WriteString(fileSeparator)
		sb.WriteString("\n\n")
	}

	if node.Type == NodeTypeDir {
		for _, child := range node.Children {
			gatherFileContents(sb, child)
		}
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
