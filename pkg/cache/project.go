package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Cache TTLs
	ProjectStateTTL = 5 * time.Minute
	ProjectFilesTTL = 1 * time.Minute
	ProjectMetaTTL  = 10 * time.Minute
)

// ProjectCache handles project-specific caching
type ProjectCache struct {
	cache *Cache
}

// NewProjectCache creates a new project cache
func NewProjectCache(cache *Cache) *ProjectCache {
	return &ProjectCache{cache: cache}
}

// ProjectState represents cached project state
type ProjectState struct {
	Phase       string                 `json:"phase"`
	Status      string                 `json:"status"`
	Progress    int                    `json:"progress"`
	FileCount   int                    `json:"file_count"`
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// projectStateKey returns the cache key for project state
func projectStateKey(projectID string) string {
	return fmt.Sprintf("project:%s:state", projectID)
}

// projectFilesKey returns the cache key for project files
func projectFilesKey(projectID string) string {
	return fmt.Sprintf("project:%s:files", projectID)
}

// projectMetaKey returns the cache key for project metadata
func projectMetaKey(projectID string) string {
	return fmt.Sprintf("project:%s:meta", projectID)
}

// GetState retrieves cached project state
func (pc *ProjectCache) GetState(ctx context.Context, projectID string) (*ProjectState, error) {
	var state ProjectState
	err := pc.cache.Get(ctx, projectStateKey(projectID), &state)
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// SetState caches project state
func (pc *ProjectCache) SetState(ctx context.Context, projectID string, state *ProjectState) error {
	state.LastUpdated = time.Now()
	return pc.cache.Set(ctx, projectStateKey(projectID), state, ProjectStateTTL)
}

// InvalidateState removes project state from cache
func (pc *ProjectCache) InvalidateState(ctx context.Context, projectID string) error {
	return pc.cache.Delete(ctx, projectStateKey(projectID))
}

// GetFiles retrieves cached file list
func (pc *ProjectCache) GetFiles(ctx context.Context, projectID string) ([]string, error) {
	var files []string
	err := pc.cache.Get(ctx, projectFilesKey(projectID), &files)
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}
	return files, nil
}

// SetFiles caches file list
func (pc *ProjectCache) SetFiles(ctx context.Context, projectID string, files []string) error {
	return pc.cache.Set(ctx, projectFilesKey(projectID), files, ProjectFilesTTL)
}

// InvalidateFiles removes file list from cache
func (pc *ProjectCache) InvalidateFiles(ctx context.Context, projectID string) error {
	return pc.cache.Delete(ctx, projectFilesKey(projectID))
}

// InvalidateAll removes all cached data for a project
func (pc *ProjectCache) InvalidateAll(ctx context.Context, projectID string) error {
	return pc.cache.Delete(ctx,
		projectStateKey(projectID),
		projectFilesKey(projectID),
		projectMetaKey(projectID),
	)
}

// CacheOrFetch implements cache-aside pattern
// If cached, returns cached value. Otherwise, calls fetch function and caches result
func (pc *ProjectCache) CacheOrFetch(ctx context.Context, projectID string, fetch func() (*ProjectState, error)) (*ProjectState, error) {
	// Try cache first
	state, err := pc.GetState(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get cached state: %w", err)
	}
	if state != nil {
		return state, nil
	}

	// Cache miss - fetch from source
	state, err = fetch()
	if err != nil {
		return nil, err
	}

	// Cache the result (ignore cache errors)
	pc.SetState(ctx, projectID, state)

	return state, nil
}
