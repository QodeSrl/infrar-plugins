package orchestrator

// OrchestratorRequest is the standard input for all orchestrators
type OrchestratorRequest struct {
	Command      string                 `json:"command"`      // "generate", "validate", etc.
	Capabilities []string               `json:"capabilities"` // List of capabilities/services to include
	Context      Context                `json:"context"`      // Project context
	Credentials  map[string]interface{} `json:"credentials"`  // Provider credentials
	Parameters   map[string]interface{} `json:"parameters"`   // Custom parameters
}

// Context contains project-level information
type Context struct {
	ProjectName string            `json:"project_name"`
	Environment string            `json:"environment"`
	Region      string            `json:"region"`
	Tags        map[string]string `json:"tags"`
	Metadata    map[string]string `json:"metadata"` // Additional metadata
}

// OrchestratorResponse is the standard output from all orchestrators
type OrchestratorResponse struct {
	Success  bool              `json:"success"`
	Files    map[string]string `json:"files"`    // filename -> content
	Metadata ResponseMetadata  `json:"metadata"` // Additional information
	Error    string            `json:"error,omitempty"`
}

// ResponseMetadata contains additional information about the generation
type ResponseMetadata struct {
	ServicesIncluded []string `json:"services_included"`
	Warnings         []string `json:"warnings"`
	RequiredAPIs     []string `json:"required_apis"` // APIs that must be enabled
}
