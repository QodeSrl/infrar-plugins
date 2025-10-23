package main

import (
	"encoding/json"
	"fmt"
	"os"
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
			ServicesIncluded: []string{"cloud-run"},
			Warnings:         []string{},
			RequiredAPIs:     []string{"run.googleapis.com"},
		},
	}

	// Get base path (parent of orchestrator directory)
	execPath, _ := os.Executable()
	basePath := filepath.Dir(filepath.Dir(execPath))

	// Generate main.tf (read from template)
	mainTf, err := os.ReadFile(filepath.Join(basePath, "terraform", "main.tf"))
	if err != nil {
		respondError(fmt.Sprintf("failed to read main.tf: %v", err))
		return
	}
	response.Files["main.tf"] = string(mainTf)

	// Generate variables.tf (read from template)
	variablesTf, err := os.ReadFile(filepath.Join(basePath, "terraform", "variables.tf"))
	if err != nil {
		respondError(fmt.Sprintf("failed to read variables.tf: %v", err))
		return
	}
	response.Files["variables.tf"] = string(variablesTf)

	// Generate tfvars from template
	tfvars, err := generateTfvars(basePath, req)
	if err != nil {
		respondError(fmt.Sprintf("failed to generate tfvars: %v", err))
		return
	}
	response.Files["terraform.tfvars"] = tfvars

	// Output response
	json.NewEncoder(os.Stdout).Encode(response)
}

func generateTfvars(basePath string, req orchestrator.OrchestratorRequest) (string, error) {
	// Read tfvars template
	tmplPath := filepath.Join(basePath, "terraform", "tfvars.tmpl")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read tfvars.tmpl: %w", err)
	}

	// Parse template
	tmpl, err := template.New("tfvars").Funcs(template.FuncMap{
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
	}).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	// Start with defaults
	variables := map[string]interface{}{
		"region": req.Context.Region,
	}

	// Merge custom variables from parameters
	for k, v := range req.Parameters {
		variables[k] = v
	}

	data := map[string]interface{}{
		"ProjectName": req.Context.ProjectName,
		"Environment": req.Context.Environment,
		"Variables":   variables,
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
