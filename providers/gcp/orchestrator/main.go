package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/infrar/plugins-shared/orchestrator"
)

func main() {
	// Read input from stdin
	var req orchestrator.OrchestratorRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		respondError(fmt.Sprintf("failed to parse input: %v", err))
		return
	}

	// Process request
	switch req.Command {
	case "generate":
		generate(req)
	default:
		respondError(fmt.Sprintf("unknown command: %s", req.Command))
	}
}

func generate(req orchestrator.OrchestratorRequest) {
	response := orchestrator.OrchestratorResponse{
		Success: true,
		Files:   make(map[string]string),
		Metadata: orchestrator.ResponseMetadata{
			ServicesIncluded: []string{},
			Warnings:         []string{},
			RequiredAPIs:     []string{},
		},
	}

	// Get base path (parent of orchestrator directory)
	execPath, _ := os.Executable()
	basePath := filepath.Dir(filepath.Dir(execPath))

	// Map capabilities to categories
	categoryMap := make(map[string][]string) // category -> capabilities
	for _, cap := range req.Capabilities {
		category := mapCapabilityToCategory(cap)
		if category != "" {
			categoryMap[category] = append(categoryMap[category], cap)
		}
	}

	// Call each category orchestrator
	var mainTfParts []string
	var variablesTfParts []string
	var tfvarsParts []string

	for category := range categoryMap {
		categoryPath := filepath.Join(basePath, "services", category, "orchestrator", "orchestrate")

		// Check if category orchestrator exists
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			response.Metadata.Warnings = append(response.Metadata.Warnings,
				fmt.Sprintf("Category %s orchestrator not found", category))
			continue
		}

		// Call category orchestrator
		categoryResp, err := callCategoryOrchestrator(categoryPath, req)
		if err != nil {
			response.Metadata.Warnings = append(response.Metadata.Warnings,
				fmt.Sprintf("Failed to call %s orchestrator: %v", category, err))
			continue
		}

		if !categoryResp.Success {
			response.Metadata.Warnings = append(response.Metadata.Warnings,
				fmt.Sprintf("Category %s failed: %s", category, categoryResp.Error))
			continue
		}

		// Collect outputs
		if mainTf, ok := categoryResp.Files["main.tf"]; ok {
			mainTfParts = append(mainTfParts, mainTf)
		}
		if variablesTf, ok := categoryResp.Files["variables.tf"]; ok {
			variablesTfParts = append(variablesTfParts, variablesTf)
		}
		if tfvars, ok := categoryResp.Files["terraform.tfvars"]; ok {
			tfvarsParts = append(tfvarsParts, tfvars)
		}

		// Merge metadata
		response.Metadata.ServicesIncluded = append(response.Metadata.ServicesIncluded,
			categoryResp.Metadata.ServicesIncluded...)
		response.Metadata.RequiredAPIs = append(response.Metadata.RequiredAPIs,
			categoryResp.Metadata.RequiredAPIs...)
	}

	// Generate provider-level configuration files
	terraformConfigPath := filepath.Join(basePath, "terraform-config")

	// Generate provider.tf
	providerTf, err := renderTemplate(filepath.Join(terraformConfigPath, "provider-block.tf.tmpl"), req)
	if err != nil {
		response.Metadata.Warnings = append(response.Metadata.Warnings,
			fmt.Sprintf("Failed to generate provider.tf: %v", err))
	} else {
		response.Files["provider.tf"] = providerTf
	}

	// Generate provider-level variables.tf
	providerVariablesTf, err := renderTemplate(filepath.Join(terraformConfigPath, "variables.tf.tmpl"), req)
	if err != nil {
		response.Metadata.Warnings = append(response.Metadata.Warnings,
			fmt.Sprintf("Failed to generate provider variables: %v", err))
	} else {
		variablesTfParts = append([]string{providerVariablesTf}, variablesTfParts...)
	}

	// Generate provider-level tfvars
	providerTfvars, err := renderTemplate(filepath.Join(terraformConfigPath, "tfvars.tmpl"), req)
	if err != nil {
		response.Metadata.Warnings = append(response.Metadata.Warnings,
			fmt.Sprintf("Failed to generate provider tfvars: %v", err))
	} else {
		tfvarsParts = append([]string{providerTfvars}, tfvarsParts...)
	}

	// Combine all parts
	if len(mainTfParts) > 0 {
		response.Files["main.tf"] = strings.Join(mainTfParts, "\n\n")
	}
	if len(variablesTfParts) > 0 {
		response.Files["variables.tf"] = strings.Join(variablesTfParts, "\n\n")
	}
	if len(tfvarsParts) > 0 {
		response.Files["terraform.tfvars"] = strings.Join(tfvarsParts, "\n\n")
	}

	// Output response
	json.NewEncoder(os.Stdout).Encode(response)
}

func mapCapabilityToCategory(capability string) string {
	// Map capabilities to their respective categories
	categoryMapping := map[string]string{
		"storage": "storage",
		"compute": "compute",
	}

	if category, ok := categoryMapping[capability]; ok {
		return category
	}
	return ""
}

func callCategoryOrchestrator(orchestratorPath string, req orchestrator.OrchestratorRequest) (*orchestrator.OrchestratorResponse, error) {
	// Prepare input JSON
	inputJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Execute category orchestrator
	cmd := exec.Command(orchestratorPath)
	cmd.Stdin = strings.NewReader(string(inputJSON))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("orchestrator execution failed: %w, output: %s", err, string(output))
	}

	// Parse response
	var resp orchestrator.OrchestratorResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse orchestrator response: %w, output: %s", err, string(output))
	}

	return &resp, nil
}

func renderTemplate(templatePath string, req orchestrator.OrchestratorRequest) (string, error) {
	// Read template
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Parse template with custom functions
	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(template.FuncMap{
		"tfstring": func(v interface{}) string {
			return fmt.Sprintf("\"%v\"", v)
		},
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
		"sanitize": func(s string) string {
			// Sanitize string for use in resource names
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, " ", "-")
			s = strings.ReplaceAll(s, "_", "-")
			return s
		},
		"sanitizeLabel": func(s string) string {
			// Sanitize string for use in GCP labels
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, " ", "-")
			s = strings.ReplaceAll(s, "_", "-")
			return s
		},
	}).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	// Extract project_id from credentials if available
	var projectID string

	// Try to extract from gcp_service_account_json (platform format)
	if saJSON, ok := req.Credentials["gcp_service_account_json"].(string); ok {
		var saData map[string]interface{}
		if err := json.Unmarshal([]byte(saJSON), &saData); err == nil {
			if pid, ok := saData["project_id"].(string); ok {
				projectID = pid
			}
		}
	}

	// Fallback: try nested gcp.project_id format
	if projectID == "" {
		if creds, ok := req.Credentials["gcp"].(map[string]interface{}); ok {
			if pid, ok := creds["project_id"].(string); ok {
				projectID = pid
			}
		}
	}

	data := map[string]interface{}{
		"ProjectName": req.Context.ProjectName,
		"Environment": req.Context.Environment,
		"Provider":    "gcp",
		"Variables": map[string]interface{}{
			"region": req.Context.Region,
		},
		"Metadata": map[string]interface{}{
			"project_id": projectID,
		},
		"Capabilities": []map[string]interface{}{},
	}

	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func respondError(message string) {
	response := orchestrator.OrchestratorResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(os.Stdout).Encode(response)
	os.Exit(1)
}
