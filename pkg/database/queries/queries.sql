-- name: CreateTenant :one
INSERT INTO tenants (name, slug, plan, settings)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: UpdateTenant :one
UPDATE tenants SET name = $2, plan = $3, settings = $4
WHERE id = $1
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (tenant_id, email, password_hash, name, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByTenantAndEmail :one
SELECT * FROM users WHERE tenant_id = $1 AND email = $2;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = NOW() WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: ListUsersByTenant :many
SELECT * FROM users WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (tenant_id, workflow_id, name, description, config, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1;

-- name: GetProjectByWorkflowID :one
SELECT * FROM projects WHERE workflow_id = $1;

-- name: ListProjectsByTenant :many
SELECT * FROM projects
WHERE tenant_id = $1 AND status != 'archived'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListProjectsByUser :many
SELECT * FROM projects
WHERE created_by = $1 AND status != 'archived'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountProjectsByTenant :one
SELECT COUNT(*) FROM projects WHERE tenant_id = $1 AND status != 'archived';

-- name: UpdateProjectPhase :exec
UPDATE projects SET phase = $2 WHERE id = $1;

-- name: UpdateProjectPhaseByWorkflowID :exec
UPDATE projects SET phase = $2, updated_at = NOW() WHERE workflow_id = $1;

-- name: UpdateProjectStatus :exec
UPDATE projects SET status = $2 WHERE id = $1;

-- name: ArchiveProject :exec
UPDATE projects SET status = 'archived' WHERE id = $1;

-- name: CreateArtifact :one
INSERT INTO artifacts (project_id, tenant_id, type, name, path, mime_type, size_bytes, checksum, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetArtifactByID :one
SELECT * FROM artifacts WHERE id = $1;

-- name: ListArtifactsByProject :many
SELECT * FROM artifacts
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: ListArtifactsByProjectAndType :many
SELECT * FROM artifacts
WHERE project_id = $1 AND type = $2
ORDER BY created_at DESC;

-- name: DeleteArtifact :exec
DELETE FROM artifacts WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1 AND NOT revoked AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1;

-- name: CleanupExpiredTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE;
