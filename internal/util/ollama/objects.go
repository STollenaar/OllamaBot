package ollama

import "time"

type GenerateResponse struct {
	Model              string    `json:"model"`
	Created            time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context"`
	TotalDuration      int       `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int       `json:"eval_duration"`
}

type GenerateRequest struct {
	Model  string                 `json:"model"`
	Prompt string                 `json:"prompt"`
	Format map[string]interface{} `json:"format"`
	Stream bool                   `json:"stream"`
}

// ListResponse is the response from [Client.List].
type ListResponse struct {
	Models []ListModelResponse `json:"models"`
}

// ListModelResponse is a single model description in [ListResponse].
type ListModelResponse struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details,omitempty"`
}

// ModelDetails provides details about a model.
type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type PullRequest struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"` // Deprecated: ignored
	Username string `json:"username"`           // Deprecated: ignored
	Password string `json:"password"`           // Deprecated: ignored
	Stream   *bool  `json:"stream,omitempty"`

	// Deprecated: set the model name with Model instead
	Name string `json:"name"`
}
