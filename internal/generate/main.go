package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	contractPath = "openapi/tradera-v4.json"
	metadataPath = "openapi/client-operations.json"
)

var httpMethods = map[string]bool{
	"delete": true,
	"get":    true,
	"patch":  true,
	"post":   true,
	"put":    true,
}

type operationMetadata struct {
	Method        string `json:"method"`
	Path          string `json:"path"`
	Name          string `json:"name"`
	AggregateName string `json:"aggregateName"`
	Service       string `json:"service"`
}

type target struct {
	service     string
	packageName string
	output      string
}

func main() {
	contract := readObject(contractPath)
	metadata := readMetadata(metadataPath)
	operations := collectOperations(contract)
	validateCoverage(operations, metadata)

	targets := []target{
		{packageName: "rest", output: "generated/rest/client.gen.go"},
		{service: "Search", packageName: "search", output: "generated/rest/search/client.gen.go"},
		{service: "Public", packageName: "public", output: "generated/rest/public/client.gen.go"},
		{service: "Listing", packageName: "listing", output: "generated/rest/listing/client.gen.go"},
		{service: "Restricted", packageName: "restricted", output: "generated/rest/restricted/client.gen.go"},
		{service: "Order", packageName: "order", output: "generated/rest/order/client.gen.go"},
		{service: "Buyer", packageName: "buyer", output: "generated/rest/buyer/client.gen.go"},
	}

	temporaryDirectory, err := os.MkdirTemp("", "tradera-go-generate-")
	check(err)
	defer os.RemoveAll(temporaryDirectory)

	for _, generationTarget := range targets {
		specification := cloneObject(contract)
		applyOperations(specification, metadata, generationTarget.service)
		specificationPath := filepath.Join(temporaryDirectory, generationTarget.packageName+".json")
		writeJSON(specificationPath, specification)
		check(os.MkdirAll(filepath.Dir(generationTarget.output), 0o755))

		command := exec.Command(
			"go", "tool", "oapi-codegen",
			"-generate", "types,client",
			"-response-type-suffix", "HTTPResponse",
			"-package", generationTarget.packageName,
			"-o", generationTarget.output,
			specificationPath,
		)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		check(command.Run())
	}
}

func readObject(path string) map[string]any {
	data, err := os.ReadFile(path)
	check(err)
	var value map[string]any
	check(json.Unmarshal(data, &value))
	return value
}

func readMetadata(path string) []operationMetadata {
	data, err := os.ReadFile(path)
	check(err)
	var metadata []operationMetadata
	check(json.Unmarshal(data, &metadata))
	return metadata
}

func collectOperations(contract map[string]any) map[string]bool {
	operations := make(map[string]bool)
	paths := contract["paths"].(map[string]any)
	for path, rawPathItem := range paths {
		if !strings.HasPrefix(path, "/v4/") {
			continue
		}
		pathItem := rawPathItem.(map[string]any)
		for method := range pathItem {
			if httpMethods[method] {
				operations[method+" "+path] = true
			}
		}
	}
	return operations
}

func validateCoverage(operations map[string]bool, metadata []operationMetadata) {
	seen := make(map[string]bool)
	for _, operation := range metadata {
		key := strings.ToLower(operation.Method) + " " + operation.Path
		if seen[key] {
			panic(fmt.Sprintf("duplicate operation metadata: %s", key))
		}
		if !operations[key] {
			panic(fmt.Sprintf("operation metadata does not exist in contract: %s", key))
		}
		seen[key] = true
	}

	var missing []string
	for operation := range operations {
		if !seen[operation] {
			missing = append(missing, operation)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		panic("missing operation metadata: " + strings.Join(missing, ", "))
	}
}

func applyOperations(contract map[string]any, metadata []operationMetadata, service string) {
	selected := make(map[string]operationMetadata)
	for _, operation := range metadata {
		if service == "" || operation.Service == service {
			selected[strings.ToLower(operation.Method)+" "+operation.Path] = operation
		}
	}

	paths := contract["paths"].(map[string]any)
	for path, rawPathItem := range paths {
		pathItem := rawPathItem.(map[string]any)
		for method, rawOperation := range pathItem {
			if !httpMethods[method] {
				continue
			}
			operation, ok := selected[method+" "+path]
			if !ok {
				delete(pathItem, method)
				continue
			}

			name := operation.Name
			if service == "" && operation.AggregateName != "" {
				name = operation.AggregateName
			}
			rawOperation.(map[string]any)["operationId"] = name
		}

		if !hasOperation(pathItem) {
			delete(paths, path)
		}
	}
}

func hasOperation(pathItem map[string]any) bool {
	for method := range pathItem {
		if httpMethods[method] {
			return true
		}
	}
	return false
}

func cloneObject(value map[string]any) map[string]any {
	data, err := json.Marshal(value)
	check(err)
	var clone map[string]any
	check(json.Unmarshal(data, &clone))
	return clone
}

func writeJSON(path string, value map[string]any) {
	data, err := json.Marshal(value)
	check(err)
	check(os.WriteFile(path, data, 0o644))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
