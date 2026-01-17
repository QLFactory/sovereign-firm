package queries_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	pgtypes "github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qlfactory/sovereign-firm/pkg/database/queries"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	// Setup: Connect to database
	ctx := context.Background()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://sovereign:sovereign123@localhost:5432/sovereign_firm?sslmode=disable"
	}

	var err error
	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testPool.Close()
	os.Exit(code)
}

func randomString(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:n]
}

func TestCreateTenant(t *testing.T) {
	ctx := context.Background()
	q := queries.New(testPool)

	slug := "test-tenant-" + randomString(8)

	tenant, err := q.CreateTenant(ctx, queries.CreateTenantParams{
		Name: "Test Tenant",
		Slug: slug,
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	if tenant.Name != "Test Tenant" {
		t.Errorf("Expected name 'Test Tenant', got '%s'", tenant.Name)
	}
	if tenant.Slug != slug {
		t.Errorf("Expected slug '%s', got '%s'", slug, tenant.Slug)
	}

	t.Logf("Created tenant: ID=%s, Name=%s, Slug=%s",
		tenant.ID.Bytes, tenant.Name, tenant.Slug)

	// Cleanup
	_, err = testPool.Exec(ctx, "DELETE FROM tenants WHERE slug = $1", slug)
	if err != nil {
		t.Logf("Warning: Failed to cleanup tenant: %v", err)
	}
}

func TestCreateUserAndAuthenticate(t *testing.T) {
	ctx := context.Background()
	q := queries.New(testPool)

	// Create tenant first
	slug := "test-tenant-" + randomString(8)
	tenant, err := q.CreateTenant(ctx, queries.CreateTenantParams{
		Name: "Auth Test Tenant",
		Slug: slug,
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Create user
	email := fmt.Sprintf("test-%s@example.com", randomString(8))
	user, err := q.CreateUser(ctx, queries.CreateUserParams{
		TenantID:     tenant.ID,
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$fakehash",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, user.Email)
	}

	t.Logf("Created user: ID=%s, Email=%s, TenantID=%s",
		user.ID.Bytes, user.Email, user.TenantID.Bytes)

	// Test GetUserByEmail
	foundUser, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if foundUser.ID != user.ID {
		t.Errorf("User IDs don't match")
	}

	t.Logf("Found user by email: %s", foundUser.Email)

	// Cleanup
	_, _ = testPool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
	_, _ = testPool.Exec(ctx, "DELETE FROM tenants WHERE slug = $1", slug)
}

func TestCreateProject(t *testing.T) {
	ctx := context.Background()
	q := queries.New(testPool)

	// Create tenant
	slug := "test-tenant-" + randomString(8)
	tenant, err := q.CreateTenant(ctx, queries.CreateTenantParams{
		Name: "Project Test Tenant",
		Slug: slug,
	})
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Create user
	email := fmt.Sprintf("test-%s@example.com", randomString(8))
	user, err := q.CreateUser(ctx, queries.CreateUserParams{
		TenantID:     tenant.ID,
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$fakehash",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create project
	workflowID := "workflow-" + randomString(8)
	project, err := q.CreateProject(ctx, queries.CreateProjectParams{
		TenantID:   tenant.ID,
		WorkflowID: workflowID,
		Name:       "Test Project",
		CreatedBy:  user.ID,
		Config:     []byte(`{"frontend": "react", "backend": "go"}`),
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", project.Name)
	}
	if project.WorkflowID != workflowID {
		t.Errorf("Expected workflow ID '%s', got '%s'", workflowID, project.WorkflowID)
	}

	t.Logf("Created project: ID=%s, Name=%s, WorkflowID=%s",
		project.ID.Bytes, project.Name, project.WorkflowID)

	// Test GetProjectByWorkflowID
	foundProject, err := q.GetProjectByWorkflowID(ctx, workflowID)
	if err != nil {
		t.Fatalf("Failed to get project by workflow ID: %v", err)
	}

	if foundProject.ID != project.ID {
		t.Errorf("Project IDs don't match")
	}

	// Test UpdateProjectPhase
	err = q.UpdateProjectPhase(ctx, queries.UpdateProjectPhaseParams{
		ID:    project.ID,
		Phase: strPtr("DEVELOPMENT"),
	})
	if err != nil {
		t.Fatalf("Failed to update project phase: %v", err)
	}

	// Verify phase update
	updatedProject, err := q.GetProjectByWorkflowID(ctx, workflowID)
	if err != nil {
		t.Fatalf("Failed to get updated project: %v", err)
	}

	if updatedProject.Phase == nil || *updatedProject.Phase != "DEVELOPMENT" {
		t.Errorf("Expected phase 'DEVELOPMENT', got '%v'", updatedProject.Phase)
	}

	t.Logf("Updated project phase to: %s", *updatedProject.Phase)

	// Cleanup
	_, _ = testPool.Exec(ctx, "DELETE FROM projects WHERE workflow_id = $1", workflowID)
	_, _ = testPool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
	_, _ = testPool.Exec(ctx, "DELETE FROM tenants WHERE slug = $1", slug)
}

func TestRefreshTokens(t *testing.T) {
	ctx := context.Background()
	q := queries.New(testPool)

	// Create tenant and user
	slug := "test-tenant-" + randomString(8)
	tenant, _ := q.CreateTenant(ctx, queries.CreateTenantParams{
		Name: "Token Test Tenant",
		Slug: slug,
	})

	email := fmt.Sprintf("test-%s@example.com", randomString(8))
	user, _ := q.CreateUser(ctx, queries.CreateUserParams{
		TenantID:     tenant.ID,
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$fakehash",
	})

	// Create refresh token
	tokenHash := "hash-" + randomString(32)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	token, err := q.CreateRefreshToken(ctx, queries.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgTimestamptz(expiresAt),
	})
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	t.Logf("Created refresh token: ID=%s, ExpiresAt=%v", token.ID.Bytes, token.ExpiresAt)

	// Test GetRefreshToken
	foundToken, err := q.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		t.Fatalf("Failed to get refresh token: %v", err)
	}

	if foundToken.UserID != user.ID {
		t.Errorf("Token user IDs don't match")
	}

	// Test RevokeRefreshToken
	err = q.RevokeRefreshToken(ctx, tokenHash)
	if err != nil {
		t.Fatalf("Failed to revoke refresh token: %v", err)
	}

	// Verify revoked - should return no rows
	_, err = q.GetRefreshToken(ctx, tokenHash)
	if err != pgx.ErrNoRows {
		t.Errorf("Expected ErrNoRows for revoked token, got: %v", err)
	}

	t.Log("Successfully revoked refresh token")

	// Cleanup
	_, _ = testPool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", user.ID.Bytes)
	_, _ = testPool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
	_, _ = testPool.Exec(ctx, "DELETE FROM tenants WHERE slug = $1", slug)
}

func strPtr(s string) *string {
	return &s
}

func pgTimestamptz(t time.Time) pgtypes.Timestamptz {
	return pgtypes.Timestamptz{Time: t, Valid: true}
}
