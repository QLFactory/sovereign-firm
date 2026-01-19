package activities

// ChatInput represents the input for the PM agent's chat activity
type ChatInput struct {
	ProjectID   string `json:"project_id"`
	ChatHistory string `json:"chat_history"`
}

// GenerateCodeInput represents the input for the Dev agent's code generation activity
type GenerateCodeInput struct {
	ProjectID string `json:"project_id"`
	Spec      string `json:"spec"`
}

// QAGenerateInput represents the input for the QA agent's test generation activity
type QAGenerateInput struct {
	ProjectID string            `json:"project_id"`
	Spec      string            `json:"spec"`
	CodeFiles map[string]string `json:"code_files"`
}

// IndexRepoParams represents the parameters for repo indexing
type IndexRepoParams struct {
	RepoURL   string `json:"repo_url"`
	ProjectID string `json:"project_id"`
}

// RefineCodeInput represents the input for the Dev agent's code refinement activity
type RefineCodeInput struct {
	ProjectID          string            `json:"project_id"`
	CurrentCode        map[string]string `json:"current_code"`
	ChatHistory        string            `json:"chat_history"`
	ValidationFeedback string            `json:"validation_feedback,omitempty"`
}
