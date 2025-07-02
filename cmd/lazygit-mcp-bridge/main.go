package main

import (
	"fmt"
	"os"

	"github.com/cappyzawa/lazygit-mcp-bridge/internal/client"
	"github.com/cappyzawa/lazygit-mcp-bridge/internal/server"
	"github.com/spf13/cobra"
)

var (
	// These will be set by goreleaser ldflags
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "lazygit-mcp-bridge",
	Short: "Bridge between lazygit and AI assistants using MCP",
	Long: `lazygit-mcp-bridge acts as a bridge between lazygit and AI assistants 
using the Model Context Protocol (MCP).

It can run as an MCP server or send messages from lazygit to the AI assistant.`,
	Version: buildVersion(),
}

func buildVersion() string {
	if version == "dev" {
		return fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date)
	}
	return version
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run as MCP server",
	Long: `Run as an MCP server that listens for messages from lazygit
and provides them to AI assistants through the Model Context Protocol.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return server.Run()
	},
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a message from lazygit to AI",
	Long: `Send a message from lazygit to the AI assistant.
This replaces the need for a separate shell script.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		line, _ := cmd.Flags().GetString("line")
		comment, _ := cmd.Flags().GetString("comment")
		
		if file == "" || line == "" || comment == "" {
			return fmt.Errorf("--file, --line, and --comment are required")
		}
		
		return client.Send(file, line, comment)
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(sendCmd)
	
	// Add flags to send command
	sendCmd.Flags().StringP("file", "f", "", "File path (required)")
	sendCmd.Flags().StringP("line", "l", "", "Line number or range (required)")
	sendCmd.Flags().StringP("comment", "c", "", "Comment for AI (required)")
	sendCmd.MarkFlagRequired("file")
	sendCmd.MarkFlagRequired("line")
	sendCmd.MarkFlagRequired("comment")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}