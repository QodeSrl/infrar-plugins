package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	// Determine which storage services to include
	services := []string{}
	for _, cap := range req.Capabilities {
		if cap == "storage" {
			services = append(services, "cloud-storage")
		}
	}

	// Call each service orchestrator
	var mainTfParts []string
	var variablesTfParts []string
	var tfvarsParts []string

	for _, service := range services {
		servicePath := filepath.Join(basePath, service, "orchestrator", "orchestrate")

		// Check if service orchestrator exists
		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			response.Metadata.Warnings = append(response.Metadata.Warnings,
				fmt.Sprintf("Service %s orchestrator not found", service))
			continue
		}

		// Call service orchestrator
		serviceResp, err := callServiceOrchestrator(servicePath, req)
		if err != nil {
			response.Metadata.Warnings = append(response.Metadata.Warnings,
				fmt.Sprintf("Failed to call %s orchestrator: %v", service, err))
			continue
		}

		if !serviceResp.Success {
			response.Metadata.Warnings = append(response.Metadata.Warnings,
				fmt.Sprintf("Service %s failed: %s", service, serviceResp.Error))
			continue
		}

		// Collect outputs
		if mainTf, ok := serviceResp.Files["main.tf"]; ok {
			mainTfParts = append(mainTfParts, mainTf)
		}
		if variablesTf, ok := serviceResp.Files["variables.tf"]; ok {
			variablesTfParts = append(variablesTfParts, variablesTf)
		}
		if tfvars, ok := serviceResp.Files["terraform.tfvars"]; ok {
			tfvarsParts = append(tfvarsParts, tfvars)
		}

		// Merge metadata
		response.Metadata.ServicesIncluded = append(response.Metadata.ServicesIncluded,
			serviceResp.Metadata.ServicesIncluded...)
		response.Metadata.RequiredAPIs = append(response.Metadata.RequiredAPIs,
			serviceResp.Metadata.RequiredAPIs...)
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

func callServiceOrchestrator(orchestratorPath string, req orchestrator.OrchestratorRequest) (*orchestrator.OrchestratorResponse, error) {
	// Prepare input JSON
	inputJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Execute service orchestrator
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

func respondError(message string) {
	response := orchestrator.OrchestratorResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(os.Stdout).Encode(response)
	os.Exit(1)
}
