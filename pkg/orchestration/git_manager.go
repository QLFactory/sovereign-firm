package orchestration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// BranchStatus represents the status of an agent's branch
type BranchStatus string

const (
	BranchActive    BranchStatus = "ACTIVE"
	BranchMerged    BranchStatus = "MERGED"
	BranchConflict  BranchStatus = "CONFLICT"
	BranchAbandoned BranchStatus = "ABANDONED"
)

// AgentBranch represents a branch created for an agent
type AgentBranch struct {
	Name       string       `json:"name"`
	AgentID    string       `json:"agent_id"`
	TaskID     string       `json:"task_id"`
	BaseBranch string       `json:"base_branch"`
	Status     BranchStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	MergedAt   *time.Time   `json:"merged_at,omitempty"`
	Commits    int          `json:"commits"`
	Files      []string     `json:"files"`
}

// MergeResult represents the result of a branch merge
type MergeResult struct {
	Success      bool     `json:"success"`
	BranchName   string   `json:"branch_name"`
	CommitHash   string   `json:"commit_hash,omitempty"`
	Conflicts    []string `json:"conflicts,omitempty"`
	Error        string   `json:"error,omitempty"`
	FilesChanged int      `json:"files_changed"`
}

// GitBranchManager manages Git branches for multi-agent coordination
type GitBranchManager struct {
	repoPath    string
	repo        *git.Repository
	baseBranch  string
	branches    map[string]*AgentBranch // branch name -> branch info
	mu          sync.RWMutex
}

// NewGitBranchManager creates a new Git branch manager
func NewGitBranchManager(repoPath, baseBranch string) (*GitBranchManager, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	if baseBranch == "" {
		baseBranch = "main"
	}

	return &GitBranchManager{
		repoPath:   repoPath,
		repo:       repo,
		baseBranch: baseBranch,
		branches:   make(map[string]*AgentBranch),
	}, nil
}

// CreateAgentBranch creates a new branch for an agent's work
func (m *GitBranchManager) CreateAgentBranch(agentID, taskID string) (*AgentBranch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Safely truncate IDs to max 8 characters
	agentPrefix := agentID
	if len(agentPrefix) > 8 {
		agentPrefix = agentPrefix[:8]
	}
	taskPrefix := taskID
	if len(taskPrefix) > 8 {
		taskPrefix = taskPrefix[:8]
	}
	branchName := fmt.Sprintf("agent/%s/%s", agentPrefix, taskPrefix)

	// Get the HEAD of base branch
	baseRef, err := m.repo.Reference(plumbing.NewBranchReferenceName(m.baseBranch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get base branch: %w", err)
	}

	// Create new branch reference
	newRef := plumbing.NewHashReference(
		plumbing.NewBranchReferenceName(branchName),
		baseRef.Hash(),
	)

	err = m.repo.Storer.SetReference(newRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	branch := &AgentBranch{
		Name:       branchName,
		AgentID:    agentID,
		TaskID:     taskID,
		BaseBranch: m.baseBranch,
		Status:     BranchActive,
		CreatedAt:  time.Now(),
		Files:      make([]string, 0),
	}

	m.branches[branchName] = branch
	return branch, nil
}

// CheckoutBranch checks out a branch
func (m *GitBranchManager) CheckoutBranch(branchName string) error {
	worktree, err := m.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// CommitChanges commits changes on the current branch
func (m *GitBranchManager) CommitChanges(branchName, message, authorName, authorEmail string, files []string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	worktree, err := m.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add files to staging
	for _, file := range files {
		_, err := worktree.Add(file)
		if err != nil {
			return "", fmt.Errorf("failed to add file %s: %w", file, err)
		}
	}

	// Create commit
	commit, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	// Update branch info
	if branch, exists := m.branches[branchName]; exists {
		branch.Commits++
		branch.Files = append(branch.Files, files...)
	}

	return commit.String(), nil
}

// MergeBranch merges an agent branch into the base branch
func (m *GitBranchManager) MergeBranch(ctx context.Context, branchName string) (*MergeResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := &MergeResult{
		BranchName: branchName,
	}

	branch, exists := m.branches[branchName]
	if !exists {
		result.Error = "branch not found"
		return result, fmt.Errorf("branch %s not found", branchName)
	}

	if branch.Status != BranchActive {
		result.Error = fmt.Sprintf("branch status is %s, cannot merge", branch.Status)
		return result, fmt.Errorf("%s", result.Error)
	}

	worktree, err := m.repo.Worktree()
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Checkout base branch
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(m.baseBranch),
	})
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Get branch reference
	branchRef, err := m.repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Perform merge using git command (go-git doesn't have native merge)
	// For now, we'll do a fast-forward or record the intent
	// In production, you'd use exec.Command for actual git merge

	// Check if fast-forward is possible
	baseRef, err := m.repo.Reference(plumbing.NewBranchReferenceName(m.baseBranch), true)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// For simplicity, assume fast-forward is possible if base hasn't changed
	// In reality, you'd check merge-base
	if baseRef.Hash() == branchRef.Hash() {
		result.Success = true
		result.CommitHash = branchRef.Hash().String()
		branch.Status = BranchMerged
		now := time.Now()
		branch.MergedAt = &now
		return result, nil
	}

	// Attempt fast-forward by updating base branch reference
	// This is a simplified merge - real implementation would handle conflicts
	newRef := plumbing.NewHashReference(
		plumbing.NewBranchReferenceName(m.baseBranch),
		branchRef.Hash(),
	)
	err = m.repo.Storer.SetReference(newRef)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.CommitHash = branchRef.Hash().String()
	result.FilesChanged = len(branch.Files)
	branch.Status = BranchMerged
	now := time.Now()
	branch.MergedAt = &now

	return result, nil
}

// GetBranchStatus returns the status of a branch
func (m *GitBranchManager) GetBranchStatus(branchName string) (*AgentBranch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	branch, exists := m.branches[branchName]
	if !exists {
		return nil, fmt.Errorf("branch %s not found", branchName)
	}
	return branch, nil
}

// ListAgentBranches returns all agent branches
func (m *GitBranchManager) ListAgentBranches() []*AgentBranch {
	m.mu.RLock()
	defer m.mu.RUnlock()

	branches := make([]*AgentBranch, 0, len(m.branches))
	for _, branch := range m.branches {
		branches = append(branches, branch)
	}
	return branches
}

// DeleteBranch deletes a branch
func (m *GitBranchManager) DeleteBranch(branchName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(branchName))
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	delete(m.branches, branchName)
	return nil
}

// CleanupMergedBranches removes branches that have been merged
func (m *GitBranchManager) CleanupMergedBranches() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for name, branch := range m.branches {
		if branch.Status == BranchMerged || branch.Status == BranchAbandoned {
			m.repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(name))
			delete(m.branches, name)
			count++
		}
	}
	return count
}

// GetConflictingFiles checks for potential conflicts between branches
func (m *GitBranchManager) GetConflictingFiles(branch1, branch2 string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b1, exists := m.branches[branch1]
	if !exists {
		return nil, fmt.Errorf("branch %s not found", branch1)
	}

	b2, exists := m.branches[branch2]
	if !exists {
		return nil, fmt.Errorf("branch %s not found", branch2)
	}

	// Find overlapping files
	files1 := make(map[string]bool)
	for _, f := range b1.Files {
		files1[f] = true
	}

	var conflicts []string
	for _, f := range b2.Files {
		if files1[f] {
			conflicts = append(conflicts, f)
		}
	}

	return conflicts, nil
}

// AbandonBranch marks a branch as abandoned
func (m *GitBranchManager) AbandonBranch(branchName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	branch, exists := m.branches[branchName]
	if !exists {
		return fmt.Errorf("branch %s not found", branchName)
	}

	branch.Status = BranchAbandoned
	return nil
}

// WriteFile writes a file to the repository on the current branch
func (m *GitBranchManager) WriteFile(filePath, content string) error {
	fullPath := filepath.Join(m.repoPath, filePath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadFile reads a file from the repository
func (m *GitBranchManager) ReadFile(filePath string) (string, error) {
	fullPath := filepath.Join(m.repoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// GetBranchDiff returns the diff between a branch and its base
func (m *GitBranchManager) GetBranchDiff(branchName string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	branch, exists := m.branches[branchName]
	if !exists {
		return "", fmt.Errorf("branch %s not found", branchName)
	}

	// Get commits on branch but not on base
	branchRef, err := m.repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		return "", err
	}

	baseRef, err := m.repo.Reference(plumbing.NewBranchReferenceName(branch.BaseBranch), true)
	if err != nil {
		return "", err
	}

	// Get commit objects
	branchCommit, err := m.repo.CommitObject(branchRef.Hash())
	if err != nil {
		return "", err
	}

	baseCommit, err := m.repo.CommitObject(baseRef.Hash())
	if err != nil {
		return "", err
	}

	// Get trees
	branchTree, err := branchCommit.Tree()
	if err != nil {
		return "", err
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return "", err
	}

	// Compare trees
	changes, err := baseTree.Diff(branchTree)
	if err != nil {
		return "", err
	}

	var diffBuilder strings.Builder
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}
		name := ""
		if change.From.Name != "" {
			name = change.From.Name
		} else if change.To.Name != "" {
			name = change.To.Name
		}
		diffBuilder.WriteString(fmt.Sprintf("%s: %s\n", action, name))
	}

	return diffBuilder.String(), nil
}
