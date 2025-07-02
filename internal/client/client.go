package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Message struct {
	File        string `json:"file"`
	Comment     string `json:"comment"`
	ProjectRoot string `json:"project_root"`
	Time        string `json:"time"`
}

func Send(file, comment string) error {
	// Get config directory
	configDir := getConfigDir()
	messageFile := filepath.Join(configDir, "jesseduffield/lazygit/mcp-messages.json")
	
	// Ensure directory exists
	messageDir := filepath.Dir(messageFile)
	if err := os.MkdirAll(messageDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Get current project root
	projectRoot := getCurrentProjectRoot()
	
	// Create message
	msg := Message{
		File:        file,
		Comment:     comment,
		ProjectRoot: projectRoot,
		Time:        time.Now().Format(time.RFC3339),
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(messageFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write message file: %w", err)
	}
	
	fmt.Printf("Message sent successfully for %s\n", file)
	return nil
}

func getConfigDir() string {
	// Follow XDG Base Directory specification
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig
	}
	// Fallback to ~/.config
	return filepath.Join(os.Getenv("HOME"), ".config")
}

func getCurrentProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	
	// Find .git directory
	current := cwd
	for current != "/" {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current
		}
		current = filepath.Dir(current)
	}
	
	return cwd // fallback to current directory
}