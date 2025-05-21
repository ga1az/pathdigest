package gitutil

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var KnownGitHosts = []string{
	"github.com",
	"gitlab.com",
	"bitbucket.org",
	"gitea.com",
	"codeberg.org",
}

type GitURLParts struct {
	RepoURL  string
	Host     string
	User     string
	RepoName string
	Branch   string
	Commit   string
	SubPath  string
	Type     string
	IsSSH    bool
}

var (
	sshURLRegex     = regexp.MustCompile(`^(?:ssh://)?git@([\w.-]+):([\w.-]+)/([\w.-]+?)(?:\.git)?(?:/(tree|blob)/([\w.-]+)/?(.*))?$`)
	httpURLRegex    = regexp.MustCompile(`^https?://([\w.-]+)/([\w.-]+)/([\w.-]+?)(?:\.git)?(?:/(tree|blob)/([\w.-]+)/?(.*))?$`)
	slugRegex       = regexp.MustCompile(`^([\w.-]+)/([\w.-]+)$`)
	commitHashRegex = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)
)

func IsLikelyGitURL(source string) bool {
	if strings.HasPrefix(source, "git@") || strings.HasPrefix(source, "ssh://") {
		return true
	}
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		_, err := url.Parse(source)
		if err == nil {
			return true
		}
	}
	if slugRegex.MatchString(source) {
		return true
	}
	return false
}

func ParseGitURL(sourceURL string) (*GitURLParts, error) {
	parts := &GitURLParts{SubPath: "/"}

	if sshMatch := sshURLRegex.FindStringSubmatch(sourceURL); len(sshMatch) > 0 {
		parts.IsSSH = true
		parts.Host = sshMatch[1]
		parts.User = sshMatch[2]
		parts.RepoName = sshMatch[3]
		parts.RepoURL = fmt.Sprintf("git@%s:%s/%s.git", parts.Host, parts.User, parts.RepoName)
		if len(sshMatch) > 5 {
			parts.Type = sshMatch[4]
			branchOrCommit := sshMatch[5]
			if commitHashRegex.MatchString(branchOrCommit) {
				parts.Commit = branchOrCommit
			} else {
				parts.Branch = branchOrCommit
			}
			if len(sshMatch) > 6 && sshMatch[6] != "" {
				parts.SubPath = "/" + strings.Trim(sshMatch[6], "/")
			}
		}
		return parts, nil
	}

	if !strings.HasPrefix(sourceURL, "http://") && !strings.HasPrefix(sourceURL, "https://") && strings.Contains(sourceURL, "/") {
		potentialHost := strings.Split(sourceURL, "/")[0]
		isKnownHost := false
		for _, h := range KnownGitHosts {
			if strings.EqualFold(potentialHost, h) {
				isKnownHost = true
				break
			}
		}
		if isKnownHost || strings.Contains(potentialHost, ".") {
			sourceURL = "https://" + sourceURL
		}
	}

	if httpMatch := httpURLRegex.FindStringSubmatch(sourceURL); len(httpMatch) > 0 {
		parts.Host = httpMatch[1]
		parts.User = httpMatch[2]
		parts.RepoName = httpMatch[3]
		parts.RepoURL = fmt.Sprintf("https://%s/%s/%s.git", parts.Host, parts.User, parts.RepoName)
		if len(httpMatch) > 5 {
			parts.Type = httpMatch[4]
			branchOrCommit := httpMatch[5]
			if commitHashRegex.MatchString(branchOrCommit) {
				parts.Commit = branchOrCommit
			} else {
				parts.Branch = branchOrCommit
			}
			if len(httpMatch) > 6 && httpMatch[6] != "" {
				parts.SubPath = "/" + strings.Trim(httpMatch[6], "/")
			}
		}
		return parts, nil
	}

	if slugMatch := slugRegex.FindStringSubmatch(sourceURL); len(slugMatch) > 0 {
		parts.User = slugMatch[1]
		parts.RepoName = slugMatch[2]
		parts.Host = "github.com" // Default
		parts.RepoURL = fmt.Sprintf("https://%s/%s/%s.git", parts.Host, parts.User, parts.RepoName)
		return parts, nil
	}

	return nil, fmt.Errorf("could not parse '%s' as a known Git URL format or slug", sourceURL)
}

func CheckRepoExists(repoURL string) (bool, error) {
	cmd := exec.Command("git", "ls-remote", repoURL)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("failed to check remote: %w (stderr: %s)", err, stderr.String())
	}
	return true, nil
}

func CloneRepo(repoURL, cloneDir, branch, commit string, subPath string, isBlob bool) (string, error) {
	repoName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	targetPath := filepath.Join(cloneDir, repoName)

	if err := os.MkdirAll(targetPath, 0755); err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("failed to create target directory %s: %w", targetPath, err)
	}

	gitArgs := []string{"clone"}

	isPartialClone := subPath != "/" && subPath != ""

	if isPartialClone {
		gitArgs = append(gitArgs, "--filter=blob:none", "--sparse")
	}

	if commit == "" {
		gitArgs = append(gitArgs, "--depth=1", "--single-branch")
		if branch != "" {
			gitArgs = append(gitArgs, "--branch", branch)
		}
	}

	gitArgs = append(gitArgs, repoURL, targetPath)

	cmd := exec.Command("git", gitArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed for %s: %w\nOutput: %s", repoURL, err, string(output))
	}

	if commit != "" {
		cmd = exec.Command("git", "-C", targetPath, "checkout", commit)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git checkout commit %s failed: %w\nOutput: %s", commit, err, string(output))
		}
	}

	if isPartialClone {
		sparsePathTarget := strings.TrimPrefix(subPath, "/")
		if isBlob {
			sparsePathTarget = filepath.Dir(sparsePathTarget)
			if sparsePathTarget == "." {
				sparsePathTarget = ""
			}
		}

		if sparsePathTarget != "" {
			cmd = exec.Command("git", "-C", targetPath, "sparse-checkout", "set", sparsePathTarget)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("git sparse-checkout set failed for %s: %w\nOutput: %s", sparsePathTarget, err, string(output))
			}
		}
	}

	return targetPath, nil
}

func FetchRemoteBranchList(repoURL string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", repoURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git ls-remote --heads failed for %s: %w\nOutput: %s", repoURL, err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	var branches []string
	for _, line := range lines {
		if strings.Contains(line, "refs/heads/") {
			parts := strings.Split(line, "refs/heads/")
			if len(parts) > 1 {
				branches = append(branches, strings.TrimSpace(parts[1]))
			}
		}
	}
	return branches, nil
}
