package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	contractURL  = "https://api.tradera.com/v4/swagger/v4/swagger.json"
	contractPath = "openapi/tradera-v4.json"
)

func main() {
	response, err := http.Get(contractURL)
	check(err)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("fetch contract: %s", response.Status))
	}

	var contract map[string]any
	check(json.NewDecoder(response.Body).Decode(&contract))
	if contract["openapi"] != "3.0.1" {
		panic(fmt.Sprintf("unexpected OpenAPI version: %v", contract["openapi"]))
	}
	info, ok := contract["info"].(map[string]any)
	if !ok || info["title"] != "Tradera API" {
		panic("unexpected OpenAPI document title")
	}

	data, err := json.MarshalIndent(contract, "", "  ")
	check(err)
	data = append(data, '\n')
	check(os.WriteFile(contractPath, data, 0o644))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
