// main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config struct mirrors the structure of our config.yaml file.
type Config struct {
	OllamaURL   string  `yaml:"ollama_url"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
}

// OllamaRequest defines the structure for the JSON payload sent to Ollama.
type OllamaRequest struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	Options struct {
		Temperature float64 `json:"temperature"`
	} `json:"options"`
}

// OllamaResponse defines the structure to decode the JSON response from Ollama.
type OllamaResponse struct {
	Response string `json:"response"`
}

// loadConfig reads and parses the configuration from the YAML file.
func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "git_commit_message", "config.yaml")
	
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file at %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		return nil, fmt.Errorf("could not parse yaml config: %w", err)
	}

	return &config, nil
}

// getStagedDiff executes `git diff --staged` and returns its output.
func getStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff")
	output, err := cmd.Output()
	if err != nil {
		// This can happen if git is not installed or not in a repo.
		return "", fmt.Errorf("failed to execute 'git diff': %w", err)
	}
	return string(output), nil
}

// generateCommitMessage sends the diff to Ollama and gets a commit message.
func generateCommitMessage(config *Config, diff string) (string, error) {
	// The prompt is crucial. It instructs the AI to act as an expert and provide a single-line message.
	prompt := fmt.Sprintf(
		"Based on the following git diff, generate a concise, single-line git commit message in the conventional commit format (e.g., 'feat: add user login' or 'fix: resolve race condition'). Do not include any explanation, preamble, or markdown formatting. Just the commit message itself.\n\nGit Diff:\n```diff\n%s\n```",
		diff,
	)

	// Construct the request payload
	apiRequest := OllamaRequest{
		Model:  config.Model,
		Prompt: prompt,
		Stream: false, // We want a single response, not a stream
	}
	apiRequest.Options.Temperature = config.Temperature

	// Marshal the request payload to JSON
	jsonData, err := json.Marshal(apiRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request to JSON: %w", err)
	}

	// Create the HTTP request
	ollamaAPIURL := fmt.Sprintf("%s/api/generate", strings.TrimSuffix(config.OllamaURL, "/"))
	req, err := http.NewRequest("POST", ollamaAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Execute the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama at %s: %w", config.OllamaURL, err)
	}
	defer resp.Body.Close()

	// Read and check the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned non-200 status: %s. Response: %s", resp.Status, string(body))
	}
	
	// Unmarshal the response
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal Ollama response: %w", err)
	}
	
	return ollamaResp.Response, nil
}

// cleanMessage removes unwanted characters like quotes and extra newlines.
func cleanMessage(msg string) string {
    // Trim leading/trailing whitespace
    cleaned := strings.TrimSpace(msg)
    // Some models wrap their output in quotes, so we remove them.
    cleaned = strings.Trim(cleaned, "\"`")
    // Ensure it's truly a single line by taking everything before the first newline.
    if idx := strings.Index(cleaned, "\n"); idx != -1 {
        cleaned = cleaned[:idx]
    }
    return cleaned
}


func main() {
	// 1. Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// 2. Get staged git diff
	diff, err := getStagedDiff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}

	if strings.TrimSpace(diff) == "" {
		fmt.Println("No staged changes found. Nothing to commit. 🤔")
		os.Exit(0)
	}

	// 3. Generate the commit message
	fmt.Println("🤖 Generating commit message from diff...")
	message, err := generateCommitMessage(config, diff)
	if err != nil {
		log.Fatalf("Error generating commit message: %v", err)
	}

	// 4. Clean and print the final message
	finalMessage := cleanMessage(message)
	fmt.Println("\n✅ Suggested Commit Message:")
	fmt.Println(finalMessage)
}
