package activities

import (
	"context"
	"log"

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
	WorkflowID string `json:"workflow_id"`
	Phase      string `json:"phase"`
}

// UpdatePhase updates the project phase in the database
func (a *ProjectAgent) UpdatePhase(ctx context.Context, params UpdatePhaseParams) error {
	if a.pool == nil {
		log.Printf("ProjectAgent: No database connection, skipping phase update for %s to %s", params.WorkflowID, params.Phase)
		return nil // Gracefully handle missing DB
	}

	q := queries.New(a.pool)
	phase := params.Phase
	err := q.UpdateProjectPhaseByWorkflowID(ctx, queries.UpdateProjectPhaseByWorkflowIDParams{
		WorkflowID: params.WorkflowID,
		Phase:      &phase,
	})
	if err != nil {
		log.Printf("ProjectAgent: Failed to update phase for %s: %v", params.WorkflowID, err)
		// Don't fail the workflow if DB update fails - it's not critical
		return nil
	}

	log.Printf("ProjectAgent: Updated phase for %s to %s", params.WorkflowID, params.Phase)
	return nil
}
