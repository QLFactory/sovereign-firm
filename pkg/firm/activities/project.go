package activities

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qlfactory/sovereign-firm/pkg/database/queries"
)

// ProjectAgent handles project-related database operations
type ProjectAgent struct {
	pool *pgxpool.Pool
}

// NewProjectAgent creates a new ProjectAgent with a database connection pool
func NewProjectAgent(pool *pgxpool.Pool) *ProjectAgent {
	return &ProjectAgent{pool: pool}
}

// UpdatePhaseParams contains parameters for updating project phase
type UpdatePhaseParams struct {
	ProjectID  string `json:"project_id"`  // UUID
	WorkflowID string `json:"workflow_id"` // Temporal ID
	Phase      string `json:"phase"`
}

// UpdatePhase updates the project phase in the database
func (a *ProjectAgent) UpdatePhase(ctx context.Context, params UpdatePhaseParams) error {
	if a.pool == nil {
		log.Printf("ProjectAgent: No database connection, skipping phase update for %s", params.WorkflowID)
		return nil
	}

	q := queries.New(a.pool)
	phase := params.Phase

	// Prefer ProjectID (UUID) if available
	if params.ProjectID != "" {
		var uuid pgtype.UUID
		if err := uuid.Scan(params.ProjectID); err == nil {
			err = q.UpdateProjectPhase(ctx, queries.UpdateProjectPhaseParams{
				ID:    uuid,
				Phase: &phase,
			})
			if err == nil {
				log.Printf("ProjectAgent: Updated phase for project %s to %s", params.ProjectID, params.Phase)
				return nil
			}
			log.Printf("ProjectAgent: Failed to update phase by ID %s: %v", params.ProjectID, err)
		}
	}

	// Fallback to WorkflowID
	err := q.UpdateProjectPhaseByWorkflowID(ctx, queries.UpdateProjectPhaseByWorkflowIDParams{
		WorkflowID: params.WorkflowID,
		Phase:      &phase,
	})
	if err != nil {
		log.Printf("ProjectAgent: Failed to update phase by WorkflowID %s: %v", params.WorkflowID, err)
		return nil
	}

	log.Printf("ProjectAgent: Updated phase for workflow %s to %s", params.WorkflowID, params.Phase)
	return nil
}
