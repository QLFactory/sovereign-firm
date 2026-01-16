package orchestration

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewFileLockManager(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)

	if mgr == nil {
		t.Fatal("manager should not be nil")
	}
}

func TestNewFileLockManager_DefaultTTL(t *testing.T) {
	mgr := NewFileLockManager(0)

	if mgr.defaultTTL != 5*time.Minute {
		t.Errorf("expected default TTL 5m, got %v", mgr.defaultTTL)
	}
}

func TestAcquireWriteLock(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	err := mgr.AcquireLock(ctx, "file.go", LockWrite, "agent-1")
	if err != nil {
		t.Fatalf("failed to acquire write lock: %v", err)
	}

	if !mgr.IsLocked("file.go") {
		t.Error("file should be locked")
	}
}

func TestAcquireReadLock(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	err := mgr.AcquireLock(ctx, "file.go", LockRead, "agent-1")
	if err != nil {
		t.Fatalf("failed to acquire read lock: %v", err)
	}

	// Multiple readers should be allowed
	err = mgr.AcquireLock(ctx, "file.go", LockRead, "agent-2")
	if err != nil {
		t.Fatalf("failed to acquire second read lock: %v", err)
	}

	if !mgr.IsLocked("file.go") {
		t.Error("file should be locked")
	}
}

func TestTryAcquireLock(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)

	success := mgr.TryAcquireLock("file.go", LockWrite, "agent-1")
	if !success {
		t.Error("should successfully acquire lock")
	}

	// Second attempt should fail
	success = mgr.TryAcquireLock("file.go", LockWrite, "agent-2")
	if success {
		t.Error("should not acquire lock held by another agent")
	}
}

func TestWriteLockBlocksOtherWrites(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)

	mgr.TryAcquireLock("file.go", LockWrite, "agent-1")

	success := mgr.TryAcquireLock("file.go", LockWrite, "agent-2")
	if success {
		t.Error("write lock should block other writes")
	}
}

func TestWriteLockBlocksReads(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)

	mgr.TryAcquireLock("file.go", LockWrite, "agent-1")

	success := mgr.TryAcquireLock("file.go", LockRead, "agent-2")
	if success {
		t.Error("write lock should block reads from other agents")
	}
}

func TestReadLockBlocksWrites(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)

	mgr.TryAcquireLock("file.go", LockRead, "agent-1")

	success := mgr.TryAcquireLock("file.go", LockWrite, "agent-2")
	if success {
		t.Error("read lock should block writes from other agents")
	}
}

func TestReleaseLock(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file.go", LockWrite, "agent-1")

	err := mgr.ReleaseLock("file.go", "agent-1")
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	if mgr.IsLocked("file.go") {
		t.Error("file should not be locked after release")
	}
}

func TestReleaseLock_WrongHolder(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file.go", LockWrite, "agent-1")

	err := mgr.ReleaseLock("file.go", "agent-2")
	if err == nil {
		t.Error("should fail to release lock held by different agent")
	}
}

func TestReleaseReadLock(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file.go", LockRead, "agent-1")
	mgr.AcquireLock(ctx, "file.go", LockRead, "agent-2")

	// Release one read lock
	err := mgr.ReleaseLock("file.go", "agent-1")
	if err != nil {
		t.Fatalf("failed to release read lock: %v", err)
	}

	// File should still be locked by agent-2
	if !mgr.IsLocked("file.go") {
		t.Error("file should still be locked")
	}

	// Release second read lock
	err = mgr.ReleaseLock("file.go", "agent-2")
	if err != nil {
		t.Fatalf("failed to release second read lock: %v", err)
	}

	// Now file should be unlocked
	if mgr.IsLocked("file.go") {
		t.Error("file should not be locked after all readers release")
	}
}

func TestReleaseAllLocks(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file1.go", LockWrite, "agent-1")
	mgr.AcquireLock(ctx, "file2.go", LockRead, "agent-1")
	mgr.AcquireLock(ctx, "file3.go", LockWrite, "agent-1")

	released := mgr.ReleaseAllLocks("agent-1")
	if released != 3 {
		t.Errorf("expected 3 released locks, got %d", released)
	}

	if len(mgr.GetLockedFiles()) != 0 {
		t.Error("all files should be unlocked")
	}
}

func TestGetLockInfo(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file.go", LockWrite, "agent-1")

	writeLock, readers := mgr.GetLockInfo("file.go")

	if writeLock == nil {
		t.Error("expected write lock info")
	}
	if writeLock.HolderID != "agent-1" {
		t.Errorf("expected holder 'agent-1', got '%s'", writeLock.HolderID)
	}
	if len(readers) != 0 {
		t.Error("expected no readers")
	}
}

func TestGetLockInfo_ReadLocks(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file.go", LockRead, "agent-1")
	mgr.AcquireLock(ctx, "file.go", LockRead, "agent-2")

	writeLock, readers := mgr.GetLockInfo("file.go")

	if writeLock != nil {
		t.Error("expected no write lock")
	}
	if len(readers) != 2 {
		t.Errorf("expected 2 readers, got %d", len(readers))
	}
}

func TestGetLockedFiles(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file1.go", LockWrite, "agent-1")
	mgr.AcquireLock(ctx, "file2.go", LockRead, "agent-1")

	locked := mgr.GetLockedFiles()

	if len(locked) != 2 {
		t.Errorf("expected 2 locked files, got %d", len(locked))
	}
}

func TestExtendLock(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file.go", LockWrite, "agent-1")

	writeLock, _ := mgr.GetLockInfo("file.go")
	originalExpiry := writeLock.ExpiresAt

	time.Sleep(10 * time.Millisecond)

	err := mgr.ExtendLock("file.go", "agent-1", 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to extend lock: %v", err)
	}

	writeLock, _ = mgr.GetLockInfo("file.go")
	if !writeLock.ExpiresAt.After(originalExpiry) {
		t.Error("expiry should be extended")
	}
}

func TestExtendLock_WrongHolder(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	mgr.AcquireLock(ctx, "file.go", LockWrite, "agent-1")

	err := mgr.ExtendLock("file.go", "agent-2", 10*time.Minute)
	if err == nil {
		t.Error("should fail to extend lock held by different agent")
	}
}

func TestAcquireMultiple(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	files := []string{"file1.go", "file2.go", "file3.go"}
	err := mgr.AcquireMultiple(ctx, files, LockWrite, "agent-1")
	if err != nil {
		t.Fatalf("failed to acquire multiple locks: %v", err)
	}

	for _, file := range files {
		if !mgr.IsLocked(file) {
			t.Errorf("file %s should be locked", file)
		}
	}
}

func TestAcquireMultiple_PartialFailure(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	// Pre-lock one file
	mgr.AcquireLock(ctx, "file2.go", LockWrite, "agent-2")

	// Use a timeout context so the test doesn't block forever
	ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	files := []string{"file1.go", "file2.go", "file3.go"}
	err := mgr.AcquireMultiple(ctxTimeout, files, LockWrite, "agent-1")
	if err == nil {
		t.Error("should fail when one file is already locked")
	}

	// All locks should be released on failure
	if mgr.TryAcquireLock("file1.go", LockWrite, "agent-3") == false {
		t.Error("file1.go should be unlocked after failure")
	}
}

func TestReleaseMultiple(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	files := []string{"file1.go", "file2.go", "file3.go"}
	mgr.AcquireMultiple(ctx, files, LockWrite, "agent-1")

	mgr.ReleaseMultiple(files, "agent-1")

	for _, file := range files {
		if mgr.IsLocked(file) {
			t.Errorf("file %s should be unlocked", file)
		}
	}
}

func TestLockUpgrade(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	// Acquire read lock
	mgr.AcquireLock(ctx, "file.go", LockRead, "agent-1")

	// Same agent should be able to upgrade to write
	success := mgr.TryAcquireLock("file.go", LockWrite, "agent-1")
	if !success {
		t.Error("should be able to upgrade read to write lock")
	}
}

func TestLockUpgrade_BlockedByOtherReaders(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	// Multiple readers
	mgr.AcquireLock(ctx, "file.go", LockRead, "agent-1")
	mgr.AcquireLock(ctx, "file.go", LockRead, "agent-2")

	// agent-1 should not be able to upgrade while agent-2 is reading
	success := mgr.TryAcquireLock("file.go", LockWrite, "agent-1")
	if success {
		t.Error("should not be able to upgrade with other readers")
	}
}

func TestConcurrentLockAcquisition(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if mgr.TryAcquireLock("file.go", LockWrite, "agent-"+string(rune('0'+id))) {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful lock, got %d", successCount)
	}
}

func TestAcquireLockWithTimeout(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)

	// Pre-lock the file
	mgr.TryAcquireLock("file.go", LockWrite, "agent-1")

	// Try to acquire with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := mgr.AcquireLock(ctx, "file.go", LockWrite, "agent-2")
	if err == nil {
		t.Error("should timeout waiting for lock")
	}
}

func TestWaiterNotification(t *testing.T) {
	mgr := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	// Pre-lock the file
	mgr.TryAcquireLock("file.go", LockWrite, "agent-1")

	// Start waiter
	done := make(chan bool, 1)
	go func() {
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		err := mgr.AcquireLock(ctxTimeout, "file.go", LockWrite, "agent-2")
		done <- (err == nil)
	}()

	// Give waiter time to start
	time.Sleep(50 * time.Millisecond)

	// Release lock
	mgr.ReleaseLock("file.go", "agent-1")

	// Waiter should succeed
	select {
	case success := <-done:
		if !success {
			t.Error("waiter should have acquired lock")
		}
	case <-time.After(2 * time.Second):
		t.Error("waiter timed out")
	}
}
