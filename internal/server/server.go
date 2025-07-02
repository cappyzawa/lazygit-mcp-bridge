package server

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"  
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// MCP Protocol types
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MCPNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

type Capabilities struct {
	Resources ResourcesCapability `json:"resources"`
	Tools     ToolsCapability     `json:"tools"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe"`
	ListChanged bool `json:"listChanged"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// Message structure for storage
type LazygitMessage struct {
	File        string `json:"file"`
	Comment     string `json:"comment"`
	ProjectRoot string `json:"project_root"`
	Time        string `json:"time"`
	Hash        string `json:"hash"` // For deduplication
}

// Message queue for lazygit messages
var messageQueue []LazygitMessage
var messageFile string
var currentProjectRoot string
var subscribers []string // Track resource subscribers

func Run() error {
	// Setup message file path following XDG Base Directory spec
	configDir := getConfigDir()
	messageFile = filepath.Join(configDir, "jesseduffield/lazygit/mcp-messages.json")

	// Get current project root (where Claude Code was launched)
	currentProjectRoot = getCurrentProjectRoot()
	log.Printf("MCP server started for project: %s", currentProjectRoot)

	// Start file watcher
	go watchMessageFile()

	// Start MCP server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		handleMCPRequest(line)
	}
	
	if err := scanner.Err(); err != nil {
		return err
	}
	
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

func watchMessageFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Watch the directory (not the file directly, as it might not exist yet)
	dir := filepath.Dir(messageFile)
	err = watcher.Add(dir)
	if err != nil {
		log.Printf("Failed to watch directory: %v", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name == messageFile && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				readMessageFile()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func readMessageFile() {
	data, err := os.ReadFile(messageFile)
	if err != nil {
		log.Printf("Failed to read message file: %v", err)
		return
	}

	var rawMessage struct {
		File        string `json:"file"`
		Comment     string `json:"comment"`
		ProjectRoot string `json:"project_root"`
		Time        string `json:"time"`
	}

	if err := json.Unmarshal(data, &rawMessage); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	// Check if message is for this project
	if rawMessage.ProjectRoot != "" && rawMessage.ProjectRoot != currentProjectRoot {
		log.Printf("Message for different project: %s (current: %s)", rawMessage.ProjectRoot, currentProjectRoot)
		return
	}

	// Create hash for deduplication (content + time)
	hashInput := strings.Join([]string{rawMessage.File, rawMessage.Comment, rawMessage.Time}, "|")
	hash := sha256.Sum256([]byte(hashInput))
	hashString := hex.EncodeToString(hash[:])

	// Check for duplicates
	for _, existingMsg := range messageQueue {
		if existingMsg.Hash == hashString {
			log.Printf("Duplicate message ignored: %s", rawMessage.Comment)
			return
		}
	}

	// Create new message with hash
	message := LazygitMessage{
		File:        rawMessage.File,
		Comment:     rawMessage.Comment,
		ProjectRoot: rawMessage.ProjectRoot,
		Time:        rawMessage.Time,
		Hash:        hashString,
	}

	// Add to message queue with limit (keep last 10 messages)
	messageQueue = append(messageQueue, message)
	if len(messageQueue) > 10 {
		messageQueue = messageQueue[1:]
	}
	
	log.Printf("Message received for project: %s (queue length: %d)", currentProjectRoot, len(messageQueue))

	// Send notification to Claude
	sendNotification(fmt.Sprintf("New message from lazygit: %s", message.Comment))

	// Note: File cleanup moved to MCP tool call to allow multiple messages
}

func handleMCPRequest(line string) {
	var req MCPRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		sendError(req.ID, -32700, "Parse error")
		return
	}

	switch req.Method {
	case "initialize":
		handleInitialize(req)
	case "initialized":
		// No response needed
	case "resources/list":
		handleResourcesList(req)
	case "resources/read":
		handleResourcesRead(req)
	case "tools/list":
		handleToolsList(req)
	case "tools/call":
		handleToolsCall(req)
	case "resources/subscribe":
		handleResourcesSubscribe(req)
	case "resources/unsubscribe":
		handleResourcesUnsubscribe(req)
	default:
		sendError(req.ID, -32601, "Method not found")
	}
}

func handleInitialize(req MCPRequest) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: Capabilities{
			Resources: ResourcesCapability{
				Subscribe:   true,
				ListChanged: true,
			},
			Tools: ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "lazygit-claude-mcp",
			Version: "1.0.0",
		},
	}
	sendResponse(req.ID, result)
}

func handleResourcesList(req MCPRequest) {
	resources := []Resource{
		{
			URI:         "lazygit://messages",
			Name:        "Lazygit Messages",
			Description: "Messages from lazygit for code improvement",
			MimeType:    "text/plain",
		},
	}
	sendResponse(req.ID, map[string]interface{}{"resources": resources})
}

func handleResourcesRead(req MCPRequest) {
	params := req.Params.(map[string]interface{})
	uri := params["uri"].(string)

	if uri == "lazygit://messages" {
		if len(messageQueue) > 0 {
			// Format all messages
			var allMessages []string
			for i, msg := range messageQueue {
				formattedMessage := fmt.Sprintf("Message %d:\nFile: %s\nComment: %s\nTime: %s\n\nPlease improve this code.", 
					i+1, msg.File, msg.Comment, msg.Time)
				allMessages = append(allMessages, formattedMessage)
			}
			
			// Join all messages with separator
			finalMessage := strings.Join(allMessages, "\n" + strings.Repeat("-", 50) + "\n\n")
			
			result := map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"uri":      uri,
						"mimeType": "text/plain",
						"text":     finalMessage,
					},
				},
			}
			sendResponse(req.ID, result)
			
			// Clear message queue and cleanup file after successful retrieval
			messageQueue = []LazygitMessage{}
			if _, err := os.Stat(messageFile); err == nil {
				os.Remove(messageFile)
				log.Printf("Message file cleaned up after retrieval")
			}
		} else {
			result := map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"uri":      uri,
						"mimeType": "text/plain",
						"text":     "No new messages from lazygit",
					},
				},
			}
			sendResponse(req.ID, result)
		}
	} else {
		sendError(req.ID, -32602, "Invalid resource URI")
	}
}

func handleToolsList(req MCPRequest) {
	tools := []Tool{
		{
			Name:        "check_lazygit_messages",
			Description: "Check for new messages from lazygit",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		},
	}
	sendResponse(req.ID, map[string]interface{}{"tools": tools})
}

func handleToolsCall(req MCPRequest) {
	params := req.Params.(map[string]interface{})
	name := params["name"].(string)

	if name == "check_lazygit_messages" {
		if len(messageQueue) > 0 {
			// Format all messages
			var allMessages []string
			for i, msg := range messageQueue {
				formattedMessage := fmt.Sprintf("Message %d:\nFile: %s\nComment: %s\nTime: %s\n\nPlease improve this code.", 
					i+1, msg.File, msg.Comment, msg.Time)
				allMessages = append(allMessages, formattedMessage)
			}
			
			// Join all messages with separator
			finalMessage := strings.Join(allMessages, "\n" + strings.Repeat("-", 50) + "\n\n")
			
			result := map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": finalMessage,
					},
				},
			}
			sendResponse(req.ID, result)
			
			// Clear message queue and cleanup file after successful retrieval
			messageQueue = []LazygitMessage{}
			if _, err := os.Stat(messageFile); err == nil {
				os.Remove(messageFile)
				log.Printf("Message file cleaned up after retrieval")
			}
		} else {
			result := map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "No new messages from lazygit",
					},
				},
			}
			sendResponse(req.ID, result)
		}
	} else {
		sendError(req.ID, -32602, "Unknown tool")
	}
}

func sendResponse(id interface{}, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func sendError(id interface{}, code int, message string) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func handleResourcesSubscribe(req MCPRequest) {
	params := req.Params.(map[string]interface{})
	uri := params["uri"].(string)
	
	if uri == "lazygit://messages" {
		subscribers = append(subscribers, uri)
		sendResponse(req.ID, map[string]interface{}{})
		log.Printf("Resource subscribed: %s", uri)
	} else {
		sendError(req.ID, -32602, "Invalid resource URI")
	}
}

func handleResourcesUnsubscribe(req MCPRequest) {
	params := req.Params.(map[string]interface{})
	uri := params["uri"].(string)
	
	// Remove from subscribers
	for i, sub := range subscribers {
		if sub == uri {
			subscribers = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
	sendResponse(req.ID, map[string]interface{}{})
	log.Printf("Resource unsubscribed: %s", uri)
}

func sendNotification(message string) {
	// Send resource update notification to subscribers
	if len(subscribers) > 0 {
		notification := MCPNotification{
			JSONRPC: "2.0",
			Method:  "notifications/resources/updated",
			Params: map[string]interface{}{
				"uri":   "lazygit://messages",
				"title": "New lazygit comment received",
			},
		}
		data, _ := json.Marshal(notification)
		fmt.Println(string(data))
		log.Printf("Sent resource update notification with title")
	}
	
	// Also send a log notification for visibility
	logNotification := MCPNotification{
		JSONRPC: "2.0",
		Method:  "notifications/message",
		Params: map[string]interface{}{
			"level":   "info",
			"message": message,
		},
	}
	logData, _ := json.Marshal(logNotification)
	fmt.Println(string(logData))
}