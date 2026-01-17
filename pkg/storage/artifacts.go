package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strings"
	"time"
)

// ArtifactType represents the type of artifact
type ArtifactType string

const (
	ArtifactTypeSource   ArtifactType = "source"
	ArtifactTypeDoc      ArtifactType = "doc"
	ArtifactTypeBuild    ArtifactType = "build"
	ArtifactTypeSnapshot ArtifactType = "snapshot"
	ArtifactTypeOther    ArtifactType = "other"
)

// Artifact represents a stored artifact
type Artifact struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	TenantID    string       `json:"tenant_id"`
	Type        ArtifactType `json:"type"`
	Name        string       `json:"name"`
	Path        string       `json:"path"`
	MimeType    string       `json:"mime_type"`
	SizeBytes   int64        `json:"size_bytes"`
	Checksum    string       `json:"checksum"`
	CreatedAt   time.Time    `json:"created_at"`
}

// ArtifactStore provides artifact storage operations
type ArtifactStore struct {
	storage *Storage
}

// NewArtifactStore creates a new artifact store
func NewArtifactStore(storage *Storage) *ArtifactStore {
	return &ArtifactStore{storage: storage}
}

// buildPath constructs the storage path for an artifact
// Format: {tenant_id}/{project_id}/{type}/{name}
func (as *ArtifactStore) buildPath(tenantID, projectID string, artifactType ArtifactType, name string) string {
	return path.Join(tenantID, projectID, string(artifactType), name)
}

// Store uploads an artifact
func (as *ArtifactStore) Store(ctx context.Context, tenantID, projectID string, artifactType ArtifactType, name string, content []byte, mimeType string) (*Artifact, error) {
	// Calculate checksum
	hash := sha256.Sum256(content)
	checksum := hex.EncodeToString(hash[:])

	// Build storage path
	storagePath := as.buildPath(tenantID, projectID, artifactType, name)

	// Determine mime type if not provided
	if mimeType == "" {
		mimeType = detectMimeType(name)
	}

	// Upload to MinIO
	reader := bytes.NewReader(content)
	if err := as.storage.PutObject(ctx, storagePath, reader, int64(len(content)), mimeType); err != nil {
		return nil, fmt.Errorf("store artifact: %w", err)
	}

	return &Artifact{
		ProjectID: projectID,
		TenantID:  tenantID,
		Type:      artifactType,
		Name:      name,
		Path:      storagePath,
		MimeType:  mimeType,
		SizeBytes: int64(len(content)),
		Checksum:  checksum,
		CreatedAt: time.Now(),
	}, nil
}

// StoreStream uploads an artifact from a reader
func (as *ArtifactStore) StoreStream(ctx context.Context, tenantID, projectID string, artifactType ArtifactType, name string, reader io.Reader, size int64, mimeType string) (*Artifact, error) {
	storagePath := as.buildPath(tenantID, projectID, artifactType, name)

	if mimeType == "" {
		mimeType = detectMimeType(name)
	}

	if err := as.storage.PutObject(ctx, storagePath, reader, size, mimeType); err != nil {
		return nil, fmt.Errorf("store artifact stream: %w", err)
	}

	return &Artifact{
		ProjectID: projectID,
		TenantID:  tenantID,
		Type:      artifactType,
		Name:      name,
		Path:      storagePath,
		MimeType:  mimeType,
		SizeBytes: size,
		CreatedAt: time.Now(),
	}, nil
}

// Get retrieves an artifact's content
func (as *ArtifactStore) Get(ctx context.Context, tenantID, projectID string, artifactType ArtifactType, name string) ([]byte, error) {
	storagePath := as.buildPath(tenantID, projectID, artifactType, name)

	reader, info, err := as.storage.GetObject(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("get artifact: %w", err)
	}
	defer reader.Close()

	content := make([]byte, info.Size)
	if _, err := io.ReadFull(reader, content); err != nil {
		return nil, fmt.Errorf("read artifact: %w", err)
	}

	return content, nil
}

// GetStream retrieves an artifact as a stream
func (as *ArtifactStore) GetStream(ctx context.Context, storagePath string) (io.ReadCloser, int64, string, error) {
	reader, info, err := as.storage.GetObject(ctx, storagePath)
	if err != nil {
		return nil, 0, "", fmt.Errorf("get artifact stream: %w", err)
	}
	return reader, info.Size, info.ContentType, nil
}

// Delete removes an artifact
func (as *ArtifactStore) Delete(ctx context.Context, storagePath string) error {
	return as.storage.DeleteObject(ctx, storagePath)
}

// List lists artifacts for a project
func (as *ArtifactStore) List(ctx context.Context, tenantID, projectID string, artifactType *ArtifactType) ([]Artifact, error) {
	prefix := path.Join(tenantID, projectID)
	if artifactType != nil {
		prefix = path.Join(prefix, string(*artifactType))
	}

	objects, err := as.storage.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}

	artifacts := make([]Artifact, 0, len(objects))
	for _, obj := range objects {
		// Parse path to extract type and name
		parts := strings.Split(obj.Key, "/")
		if len(parts) < 4 {
			continue
		}

		artifacts = append(artifacts, Artifact{
			TenantID:  parts[0],
			ProjectID: parts[1],
			Type:      ArtifactType(parts[2]),
			Name:      strings.Join(parts[3:], "/"),
			Path:      obj.Key,
			MimeType:  obj.ContentType,
			SizeBytes: obj.Size,
			CreatedAt: obj.LastModified,
		})
	}

	return artifacts, nil
}

// GetPresignedURL generates a presigned download URL
func (as *ArtifactStore) GetPresignedURL(ctx context.Context, storagePath string) (string, error) {
	return as.storage.GetPresignedURL(ctx, storagePath)
}

// CreateSnapshot creates a snapshot of all project artifacts
func (as *ArtifactStore) CreateSnapshot(ctx context.Context, tenantID, projectID, snapshotName string) error {
	// List all source files
	srcPrefix := path.Join(tenantID, projectID, string(ArtifactTypeSource))
	objects, err := as.storage.ListObjects(ctx, srcPrefix)
	if err != nil {
		return fmt.Errorf("list source files: %w", err)
	}

	// Copy each to snapshot
	for _, obj := range objects {
		// Extract relative path
		relPath := strings.TrimPrefix(obj.Key, srcPrefix+"/")
		dstPath := path.Join(tenantID, projectID, string(ArtifactTypeSnapshot), snapshotName, relPath)

		if err := as.storage.CopyObject(ctx, obj.Key, dstPath); err != nil {
			return fmt.Errorf("copy %s: %w", obj.Key, err)
		}
	}

	return nil
}

// detectMimeType determines MIME type from filename
func detectMimeType(name string) string {
	ext := strings.ToLower(path.Ext(name))
	mimeTypes := map[string]string{
		".go":    "text/x-go",
		".js":    "application/javascript",
		".ts":    "application/typescript",
		".tsx":   "application/typescript",
		".jsx":   "application/javascript",
		".json":  "application/json",
		".yaml":  "application/x-yaml",
		".yml":   "application/x-yaml",
		".md":    "text/markdown",
		".html":  "text/html",
		".css":   "text/css",
		".sql":   "application/sql",
		".py":    "text/x-python",
		".rb":    "text/x-ruby",
		".java":  "text/x-java",
		".rs":    "text/x-rust",
		".sh":    "application/x-sh",
		".txt":   "text/plain",
		".xml":   "application/xml",
		".proto": "text/x-protobuf",
		".toml":  "application/toml",
		".env":   "text/plain",
		".gitignore": "text/plain",
		".dockerfile": "text/x-dockerfile",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
