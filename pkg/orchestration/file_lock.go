package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LockType defines the type of lock
type LockType string

const (
	LockRead  LockType = "READ"
	LockWrite LockType = "WRITE"
)

// FileLock represents a lock on a file
type FileLock struct {
	FilePath  string    `json:"file_path"`
	LockType  LockType  `json:"lock_type"`
	HolderID  string    `json:"holder_id"`  // Agent ID or task ID
	AcquiredAt time.Time `json:"acquired_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// FileLockManager manages file-level locking for multi-agent coordination
type FileLockManager struct {
	locks       map[string]*FileLock   // file path -> lock
	readLocks   map[string][]string    // file path -> list of reader IDs
	waiters     map[string][]chan bool // file path -> waiting channels
	defaultTTL  time.Duration
	mu          sync.Mutex
}

// NewFileLockManager creates a new file lock manager
func NewFileLockManager(defaultTTL time.Duration) *FileLockManager {
	if defaultTTL == 0 {
		defaultTTL = 5 * time.Minute
	}
	
	mgr := &FileLockManager{
		locks:      make(map[string]*FileLock),
		readLocks:  make(map[string][]string),
		waiters:    make(map[string][]chan bool),
		defaultTTL: defaultTTL,
	}
	
	// Start cleanup goroutine
	go mgr.cleanupExpiredLocks()
	
	return mgr
}

// AcquireLock attempts to acquire a lock on a file
func (m *FileLockManager) AcquireLock(ctx context.Context, filePath string, lockType LockType, holderID string) error {
	m.mu.Lock()
	
	// Check if we can acquire immediately
	if m.canAcquire(filePath, lockType, holderID) {
		m.grantLock(filePath, lockType, holderID)
		m.mu.Unlock()
		return nil
	}
	
	// Need to wait
	waitChan := make(chan bool, 1)
	m.waiters[filePath] = append(m.waiters[filePath], waitChan)
	m.mu.Unlock()
	
	// Wait for lock or context cancellation
	select {
	case <-ctx.Done():
		m.removeWaiter(filePath, waitChan)
		return ctx.Err()
	case success := <-waitChan:
		if success {
			m.mu.Lock()
			m.grantLock(filePath, lockType, holderID)
			m.mu.Unlock()
			return nil
		}
		return fmt.Errorf("failed to acquire lock on %s", filePath)
	}
}

// TryAcquireLock attempts to acquire a lock without waiting
func (m *FileLockManager) TryAcquireLock(filePath string, lockType LockType, holderID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.canAcquire(filePath, lockType, holderID) {
		m.grantLock(filePath, lockType, holderID)
		return true
	}
	return false
}

// ReleaseLock releases a lock on a file
func (m *FileLockManager) ReleaseLock(filePath string, holderID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check write lock
	if lock, exists := m.locks[filePath]; exists {
		if lock.HolderID == holderID {
			delete(m.locks, filePath)
			m.notifyWaiters(filePath)
			return nil
		}
	}
	
	// Check read locks
	readers := m.readLocks[filePath]
	for i, readerID := range readers {
		if readerID == holderID {
			m.readLocks[filePath] = append(readers[:i], readers[i+1:]...)
			if len(m.readLocks[filePath]) == 0 {
				delete(m.readLocks, filePath)
			}
			m.notifyWaiters(filePath)
			return nil
		}
	}
	
	return fmt.Errorf("no lock held by %s on %s", holderID, filePath)
}

// ReleaseAllLocks releases all locks held by a specific holder
func (m *FileLockManager) ReleaseAllLocks(holderID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	released := 0
	
	// Release write locks
	for filePath, lock := range m.locks {
		if lock.HolderID == holderID {
			delete(m.locks, filePath)
			m.notifyWaiters(filePath)
			released++
		}
	}
	
	// Release read locks
	for filePath, readers := range m.readLocks {
		var newReaders []string
		for _, readerID := range readers {
			if readerID != holderID {
				newReaders = append(newReaders, readerID)
			} else {
				released++
			}
		}
		if len(newReaders) == 0 {
			delete(m.readLocks, filePath)
			m.notifyWaiters(filePath)
		} else {
			m.readLocks[filePath] = newReaders
		}
	}
	
	return released
}

// GetLockInfo returns information about a lock on a file
func (m *FileLockManager) GetLockInfo(filePath string) (*FileLock, []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	writeLock := m.locks[filePath]
	readLocks := make([]string, len(m.readLocks[filePath]))
	copy(readLocks, m.readLocks[filePath])
	
	return writeLock, readLocks
}

// IsLocked returns true if the file has any lock
func (m *FileLockManager) IsLocked(filePath string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	_, hasWrite := m.locks[filePath]
	readers := m.readLocks[filePath]
	return hasWrite || len(readers) > 0
}

// GetLockedFiles returns all currently locked files
func (m *FileLockManager) GetLockedFiles() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	files := make(map[string]bool)
	for path := range m.locks {
		files[path] = true
	}
	for path := range m.readLocks {
		files[path] = true
	}
	
	result := make([]string, 0, len(files))
	for path := range files {
		result = append(result, path)
	}
	return result
}

// canAcquire checks if a lock can be acquired (must hold mutex)
func (m *FileLockManager) canAcquire(filePath string, lockType LockType, holderID string) bool {
	writeLock, hasWrite := m.locks[filePath]
	readers := m.readLocks[filePath]
	
	if lockType == LockWrite {
		// Write lock requires no other locks
		if hasWrite && writeLock.HolderID != holderID {
			return false
		}
		if len(readers) > 0 {
			// Allow upgrade if we're the only reader
			if len(readers) == 1 && readers[0] == holderID {
				return true
			}
			return false
		}
		return true
	}
	
	// Read lock
	if hasWrite && writeLock.HolderID != holderID {
		return false
	}
	return true
}

// grantLock grants a lock (must hold mutex)
func (m *FileLockManager) grantLock(filePath string, lockType LockType, holderID string) {
	now := time.Now()
	
	if lockType == LockWrite {
		// Remove any read lock we might have
		readers := m.readLocks[filePath]
		var newReaders []string
		for _, r := range readers {
			if r != holderID {
				newReaders = append(newReaders, r)
			}
		}
		if len(newReaders) == 0 {
			delete(m.readLocks, filePath)
		} else {
			m.readLocks[filePath] = newReaders
		}
		
		m.locks[filePath] = &FileLock{
			FilePath:   filePath,
			LockType:   LockWrite,
			HolderID:   holderID,
			AcquiredAt: now,
			ExpiresAt:  now.Add(m.defaultTTL),
		}
	} else {
		m.readLocks[filePath] = append(m.readLocks[filePath], holderID)
	}
}

// notifyWaiters signals waiting goroutines (must hold mutex)
func (m *FileLockManager) notifyWaiters(filePath string) {
	waiters := m.waiters[filePath]
	if len(waiters) > 0 {
		// Notify first waiter
		waiters[0] <- true
		m.waiters[filePath] = waiters[1:]
		if len(m.waiters[filePath]) == 0 {
			delete(m.waiters, filePath)
		}
	}
}

// removeWaiter removes a waiter from the queue
func (m *FileLockManager) removeWaiter(filePath string, waitChan chan bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	waiters := m.waiters[filePath]
	for i, ch := range waiters {
		if ch == waitChan {
			m.waiters[filePath] = append(waiters[:i], waiters[i+1:]...)
			break
		}
	}
	if len(m.waiters[filePath]) == 0 {
		delete(m.waiters, filePath)
	}
}

// cleanupExpiredLocks periodically removes expired locks
func (m *FileLockManager) cleanupExpiredLocks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		
		for filePath, lock := range m.locks {
			if now.After(lock.ExpiresAt) {
				delete(m.locks, filePath)
				m.notifyWaiters(filePath)
			}
		}
		
		m.mu.Unlock()
	}
}

// ExtendLock extends the expiration time of an existing lock
func (m *FileLockManager) ExtendLock(filePath string, holderID string, extension time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	lock, exists := m.locks[filePath]
	if !exists || lock.HolderID != holderID {
		return fmt.Errorf("no write lock held by %s on %s", holderID, filePath)
	}
	
	lock.ExpiresAt = time.Now().Add(extension)
	return nil
}

// AcquireMultiple acquires locks on multiple files atomically
func (m *FileLockManager) AcquireMultiple(ctx context.Context, files []string, lockType LockType, holderID string) error {
	// Try to acquire all locks
	acquired := make([]string, 0, len(files))
	
	for _, filePath := range files {
		err := m.AcquireLock(ctx, filePath, lockType, holderID)
		if err != nil {
			// Release all acquired locks
			for _, f := range acquired {
				m.ReleaseLock(f, holderID)
			}
			return fmt.Errorf("failed to acquire lock on %s: %w", filePath, err)
		}
		acquired = append(acquired, filePath)
	}
	
	return nil
}

// ReleaseMultiple releases locks on multiple files
func (m *FileLockManager) ReleaseMultiple(files []string, holderID string) {
	for _, filePath := range files {
		m.ReleaseLock(filePath, holderID)
	}
}
