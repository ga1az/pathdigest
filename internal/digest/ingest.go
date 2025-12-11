package digest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ga1az/pathdigest/internal/fsutil"
	"github.com/ga1az/pathdigest/internal/gitutil"
)

const (
	maxDepth = 20
)

func ProcessSource(opts IngestionOptions) (*Result, error) {
	if gitutil.IsLikelyGitURL(opts.Source) {
		return processGitURL(opts)
	}
	return processLocalPath(opts)
}

func processLocalPath(opts IngestionOptions) (*Result, error) {
	absSourcePath, err := filepath.Abs(opts.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for source: %w", err)
	}

	info, err := os.Stat(absSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source path does not exist: %s", absSourcePath)
		}
		return nil, fmt.Errorf("failed to stat source path %s: %w", absSourcePath, err)
	}

	rootNode := &FileNode{
		Name:     filepath.Base(absSourcePath),
		Path:     ".", // For the root node, the relative path to itself is "."
		FullPath: absSourcePath,
		Mode:     info.Mode(),
		Depth:    0,
	}

	var totalFilesIngested int
	var totalSizeIngested int64

	if info.IsDir() {
		rootNode.Type = NodeTypeDir
		err = processDirectory(rootNode, absSourcePath, opts, &totalFilesIngested, &totalSizeIngested, 0)
		if err != nil {
			return nil, err
		}
	} else { // It's a single file
		rootNode.Type = NodeTypeFile
		rootNode.Size = info.Size()

		relPathForPattern := info.Name()

		matchesExclude := isPathMatchWithInfo(relPathForPattern, false, opts.ExcludePatterns)
		matchesInclude := false
		if len(opts.IncludePatterns) > 0 {
			matchesInclude = isPathMatchWithInfo(relPathForPattern, false, opts.IncludePatterns)
		}

		var finalDecisionToProcess bool
		if matchesExclude {
			if len(opts.IncludePatterns) > 0 && matchesInclude {
				finalDecisionToProcess = true
			} else {
				finalDecisionToProcess = false
			}
		} else {
			if len(opts.IncludePatterns) > 0 {
				finalDecisionToProcess = matchesInclude
			} else {
				finalDecisionToProcess = true
			}
		}

		if !finalDecisionToProcess {
			rootNode.Type = NodeTypeExcluded
		} else if opts.MaxFileSize > 0 && rootNode.Size > opts.MaxFileSize {
			rootNode.Type = NodeTypeTooLarge
			totalFilesIngested = 1
			totalSizeIngested = rootNode.Size
		} else {
			isText, errText := fsutil.IsTextFile(absSourcePath)
			if errText != nil {
				rootNode.Error = fmt.Errorf("error checking if file is text: %w", errText)
				rootNode.Type = NodeTypeNotText
			} else if !isText {
				rootNode.Type = NodeTypeNotText
			} else {
				// TODO: Handle .ipynb (currently read as plain text)
				content, errRead := fsutil.ReadFileContent(absSourcePath)
				if errRead != nil {
					rootNode.Error = fmt.Errorf("error reading file content: %w", errRead)
				} else {
					rootNode.Content = content
				}
			}
			totalFilesIngested = 1
			totalSizeIngested = rootNode.Size
		}
	}

	result := &Result{
		RootNode:   rootNode,
		TotalFiles: totalFilesIngested,
		TotalSize:  totalSizeIngested,
	}

	return result, nil
}

func processGitURL(opts IngestionOptions) (*Result, error) {
	fmt.Fprintf(os.Stderr, "Processing Git URL: %s\n", opts.Source)

	gitParts, err := gitutil.ParseGitURL(opts.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse git URL: %w", err)
	}

	if opts.Branch != "" { // CLI branch override
		gitParts.Branch = opts.Branch
		gitParts.Commit = "" // CLI branch overrides URL commit
	}

	fmt.Fprintf(os.Stderr, "Checking if repository %s exists...\n", gitParts.RepoURL)
	exists, err := gitutil.CheckRepoExists(gitParts.RepoURL)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("repository %s does not exist or is not accessible", gitParts.RepoURL)
	}

	tempCloneDir, err := os.MkdirTemp("", "pathdigest-clone-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary clone directory: %w", err)
	}
	defer func() {
		fmt.Fprintf(os.Stderr, "Cleaning up temporary directory: %s\n", tempCloneDir)
		os.RemoveAll(tempCloneDir)
	}()

	fmt.Fprintf(os.Stderr, "Cloning %s (branch: %s, commit: %s, subPath: %s) into %s...\n", gitParts.RepoURL, gitParts.Branch, gitParts.Commit, gitParts.SubPath, tempCloneDir)
	clonedRepoPath, err := gitutil.CloneRepo(gitParts.RepoURL, tempCloneDir, gitParts.Branch, gitParts.Commit, gitParts.SubPath, gitParts.Type == "blob")
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	pathToProcess := clonedRepoPath
	if gitParts.Type == "blob" && gitParts.SubPath != "/" && gitParts.SubPath != "" {
		pathToProcess = filepath.Join(clonedRepoPath, strings.TrimPrefix(gitParts.SubPath, "/"))
	}

	localOpts := opts
	localOpts.Source = pathToProcess

	ingestResult, err := processLocalPath(localOpts)
	if err != nil {
		return nil, err
	}

	if ingestResult.RootNode != nil {
		repoIdentifier := fmt.Sprintf("%s/%s", gitParts.User, gitParts.RepoName)
		if gitParts.Host != "" && !strings.Contains(repoIdentifier, gitParts.Host) {
		}

		fullIdentifier := repoIdentifier
		if gitParts.SubPath != "/" && gitParts.SubPath != "" && gitParts.Type != "blob" {
			fullIdentifier += strings.TrimSuffix(gitParts.SubPath, "/")
		} else if gitParts.Type == "blob" {
			// TODO: Prefix with repo path
		}
		ingestResult.RootNode.Name = fullIdentifier
		ingestResult.GitInfo = gitParts
	}

	return ingestResult, nil
}

func processDirectory(currentDirNode *FileNode, basePath string, opts IngestionOptions, totalFiles *int, totalSize *int64, currentDepth int) error {
	if currentDepth >= maxDepth {
		fmt.Fprintf(os.Stderr, "Warning: Maximum directory depth (%d) reached at %s\n", maxDepth, currentDirNode.FullPath)
		return nil
	}

	entries, err := os.ReadDir(currentDirNode.FullPath)
	if err != nil {
		currentDirNode.Error = fmt.Errorf("failed to read directory %s: %w", currentDirNode.FullPath, err)
		return nil
	}

	currentDirNode.Children = make([]*FileNode, 0, len(entries))

	for _, entry := range entries {
		entryPath := filepath.Join(currentDirNode.FullPath, entry.Name())
		relPath, errRel := fsutil.GetRelativePath(basePath, entryPath)
		if errRel != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not get relative path for %s: %v\n", entryPath, errRel)
			relPath = entry.Name()
		}

		info, errInfo := entry.Info()
		if errInfo != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not get info for %s: %v\n", entryPath, errInfo)
			childNode := &FileNode{
				Name: entry.Name(), Path: relPath, FullPath: entryPath,
				Type: NodeTypeFile, Error: fmt.Errorf("could not get file info: %w", errInfo), Depth: currentDepth + 1,
			}
			currentDirNode.Children = append(currentDirNode.Children, childNode)
			continue
		}

		matchesExclude := isPathMatchWithInfo(relPath, info.IsDir(), opts.ExcludePatterns)
		matchesInclude := false
		if len(opts.IncludePatterns) > 0 {
			matchesInclude = isPathMatchWithInfo(relPath, info.IsDir(), opts.IncludePatterns)
		}

		var finalDecisionToProcess bool

		if entry.IsDir() {
			if matchesExclude && !matchesInclude {
				continue
			}
			if len(opts.IncludePatterns) > 0 {
				finalDecisionToProcess = matchesInclude || shouldProcessDirForInclude(relPath, opts.IncludePatterns)
			} else {
				finalDecisionToProcess = true
			}
		} else {
			if matchesExclude {
				if len(opts.IncludePatterns) > 0 && matchesInclude {
					finalDecisionToProcess = true
				} else {
					finalDecisionToProcess = false
				}
			} else {
				if len(opts.IncludePatterns) > 0 {
					finalDecisionToProcess = matchesInclude
				} else {
					finalDecisionToProcess = true
				}
			}
		}

		if !finalDecisionToProcess {
			continue
		}

		childNode := &FileNode{
			Name: entry.Name(), Path: relPath, FullPath: entryPath,
			Size: info.Size(), Mode: info.Mode(), Depth: currentDepth + 1,
		}

		if entry.IsDir() {
			childNode.Type = NodeTypeDir
			currentDirNode.Children = append(currentDirNode.Children, childNode)
			errProcDir := processDirectory(childNode, basePath, opts, totalFiles, totalSize, currentDepth+1)
			if errProcDir != nil {
			}
			currentDirNode.Size += childNode.Size
		} else if info.Mode().IsRegular() {
			childNode.Type = NodeTypeFile

			if opts.MaxFileSize > 0 && childNode.Size > opts.MaxFileSize {
				childNode.Type = NodeTypeTooLarge
			} else {
				isText, errText := fsutil.IsTextFile(entryPath)
				if errText != nil {
					childNode.Error = fmt.Errorf("error checking if file is text: %w", errText)
					childNode.Type = NodeTypeNotText
				} else if !isText {
					childNode.Type = NodeTypeNotText
				} else {
					// TODO: Handle .ipynb
					content, errRead := fsutil.ReadFileContent(entryPath)
					if errRead != nil {
						childNode.Error = fmt.Errorf("error reading file content: %w", errRead)
					} else {
						childNode.Content = content
					}
				}
			}
			*totalFiles++
			*totalSize += childNode.Size
			currentDirNode.Size += childNode.Size
			currentDirNode.Children = append(currentDirNode.Children, childNode)
		} else if info.Mode()&fs.ModeSymlink != 0 {
			childNode.Type = NodeTypeSymlink
			// TODO: Read symlink destination and save in FileNode
			// target, _ := os.Readlink(entryPath)
			// childNode.SymlinkTarget = target
			currentDirNode.Children = append(currentDirNode.Children, childNode)
		}
	}
	sortNodes(currentDirNode.Children)
	return nil
}

func isPathMatchWithInfo(relativePath string, isDir bool, patterns []string) bool {
	normalizedPath := filepath.ToSlash(relativePath)
	normalizedPath = strings.TrimPrefix(normalizedPath, "./")
	if normalizedPath == "" && relativePath == "./" {
		normalizedPath = "."
	}

	for _, pattern := range patterns {
		isDirPattern := strings.HasSuffix(pattern, "/")
		cleanPattern := strings.TrimSuffix(pattern, "/")
		cleanPattern = strings.TrimPrefix(cleanPattern, "./")

		if cleanPattern == "" && pattern == "./" {
			cleanPattern = "."
		}

		if cleanPattern == "" && isDirPattern {
			if normalizedPath == "." {
				return true
			}
			continue
		}
		if cleanPattern == "" && !isDirPattern {
			continue
		}

		if isDirPattern {
			var pathToCheckPrefix string
			if isDir {
				pathToCheckPrefix = normalizedPath + "/"
			} else {
				parentDir := filepath.Dir(normalizedPath)
				if parentDir == "." {
					parentDir = ""
				}
				pathToCheckPrefix = parentDir + "/"
			}

			if strings.HasPrefix(pathToCheckPrefix, cleanPattern+"/") {
				return true
			}

			if isDir && strings.HasPrefix(cleanPattern+"/", pathToCheckPrefix) {
				return true
			}
		} else {
			baseName := filepath.Base(normalizedPath)
			if matched, _ := filepath.Match(cleanPattern, baseName); matched {
				return true
			}
			if matched, _ := filepath.Match(cleanPattern, normalizedPath); matched {
				return true
			}
		}
	}
	return false
}

func shouldProcessDirForInclude(relativePath string, patterns []string) bool {
	normalizedPath := filepath.ToSlash(relativePath)
	normalizedPath = strings.TrimPrefix(normalizedPath, "./")

	for _, pattern := range patterns {
		isDirPattern := strings.HasSuffix(pattern, "/")
		cleanPattern := strings.TrimSuffix(pattern, "/")
		cleanPattern = strings.TrimPrefix(cleanPattern, "./")

		if isDirPattern {
			pathWithSlash := normalizedPath + "/"
			if strings.HasPrefix(pathWithSlash, cleanPattern+"/") {
				return true
			}
			if strings.HasPrefix(cleanPattern+"/", pathWithSlash) {
				return true
			}
		} else {
			return true
		}
	}
	return false
}

func sortNodes(nodes []*FileNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		nodeI := nodes[i]
		nodeJ := nodes[j]
		if nodeI.Type == NodeTypeDir && nodeJ.Type != NodeTypeDir {
			return true
		}
		if nodeI.Type != NodeTypeDir && nodeJ.Type == NodeTypeDir {
			return false
		}
		return strings.ToLower(nodeI.Name) < strings.ToLower(nodeJ.Name)
	})
}
