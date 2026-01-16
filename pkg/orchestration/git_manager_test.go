package orchestration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize git repository with main branch
	repo, err := git.PlainInitWithOptions(tmpDir, &git.PlainInitOptions{
		InitOptions: git.InitOptions{
			DefaultBranch: "refs/heads/main",
		},
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to init repo: %v", err)
	}

	// Create initial commit (git requires at least one commit for branches)
	worktree, err := repo.Worktree()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create a README file
	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo\n"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to write README: %v", err)
	}

	// Add and commit
	if _, err := worktree.Add("README.md"); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to add file: %v", err)
	}

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to commit: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestNewGitBranchManager(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	if mgr.repoPath != tmpDir {
		t.Errorf("expected repoPath %s, got %s", tmpDir, mgr.repoPath)
	}
	if mgr.baseBranch != "main" {
		t.Errorf("expected baseBranch 'main', got '%s'", mgr.baseBranch)
	}
}

func TestNewGitBranchManager_DefaultBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	if mgr.baseBranch != "main" {
		t.Errorf("expected default baseBranch 'main', got '%s'", mgr.baseBranch)
	}
}

func TestNewGitBranchManager_InvalidRepo(t *testing.T) {
	_, err := NewGitBranchManager("/nonexistent/path", "main")
	if err == nil {
		t.Error("expected error for non-existent repo")
	}
}

func TestCreateAgentBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	branch, err := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	if branch.AgentID != "agent-12345678" {
		t.Errorf("expected AgentID 'agent-12345678', got '%s'", branch.AgentID)
	}
	if branch.TaskID != "task-abcdefgh" {
		t.Errorf("expected TaskID 'task-abcdefgh', got '%s'", branch.TaskID)
	}
	if branch.Status != BranchActive {
		t.Errorf("expected status ACTIVE, got %s", branch.Status)
	}
	if branch.BaseBranch != "main" {
		t.Errorf("expected BaseBranch 'main', got '%s'", branch.BaseBranch)
	}

	// Verify branch exists (IDs truncated to 8 chars)
	expectedName := "agent/agent-12/task-abc"
	if branch.Name != expectedName {
		t.Errorf("expected branch name '%s', got '%s'", expectedName, branch.Name)
	}
}

func TestCheckoutBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	branch, err := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	err = mgr.CheckoutBranch(branch.Name)
	if err != nil {
		t.Fatalf("failed to checkout branch: %v", err)
	}

	// Checkout back to main
	err = mgr.CheckoutBranch("main")
	if err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
}

func TestCheckoutBranch_NonExistent(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = mgr.CheckoutBranch("nonexistent-branch")
	if err == nil {
		t.Error("expected error for non-existent branch")
	}
}

func TestWriteAndReadFile(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Write a file
	content := "package main\n\nfunc main() {}\n"
	err = mgr.WriteFile("src/main.go", content)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Read it back
	readContent, err := mgr.ReadFile("src/main.go")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if readContent != content {
		t.Errorf("content mismatch: expected %q, got %q", content, readContent)
	}
}

func TestWriteFile_CreatesDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Write file in nested directory
	err = mgr.WriteFile("deep/nested/path/file.txt", "test content")
	if err != nil {
		t.Fatalf("failed to write file in nested dir: %v", err)
	}

	// Verify file exists
	content, err := mgr.ReadFile("deep/nested/path/file.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if content != "test content" {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestReadFile_NonExistent(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = mgr.ReadFile("nonexistent.txt")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestCommitChanges(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create branch and checkout
	branch, err := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	mgr.CheckoutBranch(branch.Name)

	// Write a file
	err = mgr.WriteFile("new_file.go", "package main\n")
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Commit
	hash, err := mgr.CommitChanges(branch.Name, "Add new file", "Agent", "agent@test.com", []string{"new_file.go"})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty commit hash")
	}

	// Verify branch commit count updated
	branchInfo, _ := mgr.GetBranchStatus(branch.Name)
	if branchInfo.Commits != 1 {
		t.Errorf("expected 1 commit, got %d", branchInfo.Commits)
	}
}

func TestGetBranchStatus(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	branch, err := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	status, err := mgr.GetBranchStatus(branch.Name)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.Status != BranchActive {
		t.Errorf("expected ACTIVE status, got %s", status.Status)
	}
}

func TestGetBranchStatus_NotFound(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = mgr.GetBranchStatus("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent branch")
	}
}

func TestListAgentBranches(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create multiple branches
	mgr.CreateAgentBranch("agent-1", "task-1")
	mgr.CreateAgentBranch("agent-2", "task-2")
	mgr.CreateAgentBranch("agent-3", "task-3")

	branches := mgr.ListAgentBranches()

	if len(branches) != 3 {
		t.Errorf("expected 3 branches, got %d", len(branches))
	}
}

func TestDeleteBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	branch, err := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	err = mgr.DeleteBranch(branch.Name)
	if err != nil {
		t.Fatalf("failed to delete branch: %v", err)
	}

	// Verify branch is gone from manager
	branches := mgr.ListAgentBranches()
	if len(branches) != 0 {
		t.Errorf("expected 0 branches after delete, got %d", len(branches))
	}
}

func TestAbandonBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	branch, err := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	err = mgr.AbandonBranch(branch.Name)
	if err != nil {
		t.Fatalf("failed to abandon branch: %v", err)
	}

	status, _ := mgr.GetBranchStatus(branch.Name)
	if status.Status != BranchAbandoned {
		t.Errorf("expected ABANDONED status, got %s", status.Status)
	}
}

func TestAbandonBranch_NotFound(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = mgr.AbandonBranch("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent branch")
	}
}

func TestCleanupMergedBranches(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create branches with different statuses
	branch1, _ := mgr.CreateAgentBranch("agent-1", "task-1")
	branch2, _ := mgr.CreateAgentBranch("agent-2", "task-2")
	mgr.CreateAgentBranch("agent-3", "task-3") // stays active

	// Mark some as merged/abandoned
	mgr.branches[branch1.Name].Status = BranchMerged
	mgr.branches[branch2.Name].Status = BranchAbandoned

	cleaned := mgr.CleanupMergedBranches()

	if cleaned != 2 {
		t.Errorf("expected 2 cleaned branches, got %d", cleaned)
	}

	// Only active branch should remain
	branches := mgr.ListAgentBranches()
	if len(branches) != 1 {
		t.Errorf("expected 1 remaining branch, got %d", len(branches))
	}
}

func TestGetConflictingFiles(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create two branches
	branch1, _ := mgr.CreateAgentBranch("agent-1", "task-1")
	branch2, _ := mgr.CreateAgentBranch("agent-2", "task-2")

	// Simulate file modifications
	mgr.branches[branch1.Name].Files = []string{"shared.go", "unique1.go"}
	mgr.branches[branch2.Name].Files = []string{"shared.go", "unique2.go"}

	conflicts, err := mgr.GetConflictingFiles(branch1.Name, branch2.Name)
	if err != nil {
		t.Fatalf("failed to get conflicts: %v", err)
	}

	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0] != "shared.go" {
		t.Errorf("expected 'shared.go' conflict, got '%s'", conflicts[0])
	}
}

func TestGetConflictingFiles_NoConflicts(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	branch1, _ := mgr.CreateAgentBranch("agent-1", "task-1")
	branch2, _ := mgr.CreateAgentBranch("agent-2", "task-2")

	// Different files
	mgr.branches[branch1.Name].Files = []string{"file1.go"}
	mgr.branches[branch2.Name].Files = []string{"file2.go"}

	conflicts, err := mgr.GetConflictingFiles(branch1.Name, branch2.Name)
	if err != nil {
		t.Fatalf("failed to get conflicts: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts))
	}
}

func TestMergeBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create branch
	branch, err := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Checkout and add changes
	mgr.CheckoutBranch(branch.Name)
	mgr.WriteFile("new_feature.go", "package feature\n")
	mgr.CommitChanges(branch.Name, "Add feature", "Agent", "agent@test.com", []string{"new_feature.go"})

	// Merge
	ctx := context.Background()
	result, err := mgr.MergeBranch(ctx, branch.Name)
	if err != nil {
		t.Fatalf("failed to merge: %v", err)
	}

	if !result.Success {
		t.Errorf("merge should succeed: %s", result.Error)
	}

	// Check branch status updated
	status, _ := mgr.GetBranchStatus(branch.Name)
	if status.Status != BranchMerged {
		t.Errorf("expected MERGED status, got %s", status.Status)
	}
	if status.MergedAt == nil {
		t.Error("MergedAt should be set")
	}
}

func TestMergeBranch_NotFound(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	result, err := mgr.MergeBranch(ctx, "nonexistent")

	if err == nil {
		t.Error("expected error for non-existent branch")
	}
	if result.Success {
		t.Error("result should not be successful")
	}
}

func TestMergeBranch_AlreadyMerged(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	mgr, err := NewGitBranchManager(tmpDir, "main")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	branch, _ := mgr.CreateAgentBranch("agent-12345678", "task-abcdefgh")
	mgr.branches[branch.Name].Status = BranchMerged

	ctx := context.Background()
	result, err := mgr.MergeBranch(ctx, branch.Name)

	if err == nil {
		t.Error("expected error for already merged branch")
	}
	if result.Success {
		t.Error("result should not be successful")
	}
}

func TestBranchStatus_Constants(t *testing.T) {
	if BranchActive != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s", BranchActive)
	}
	if BranchMerged != "MERGED" {
		t.Errorf("expected MERGED, got %s", BranchMerged)
	}
	if BranchConflict != "CONFLICT" {
		t.Errorf("expected CONFLICT, got %s", BranchConflict)
	}
	if BranchAbandoned != "ABANDONED" {
		t.Errorf("expected ABANDONED, got %s", BranchAbandoned)
	}
}
