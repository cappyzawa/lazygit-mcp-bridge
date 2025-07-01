package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"  
	"os"
	"path/filepath"

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

// Message queue for lazygit messages
var messageQueue []string
var messageFile string

func main() {
	// Setup message file path
	configDir := os.Getenv("HOME") + "/.config/jesseduffield/lazygit"
	messageFile = filepath.Join(configDir, "claude-messages.json")

	// Start file watcher
	go watchMessageFile()

	// Start MCP server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		handleMCPRequest(line)
	}
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

	var message struct {
		File    string `json:"file"`
		Line    string `json:"line"`
		Comment string `json:"comment"`
		Time    string `json:"time"`
	}

	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	// Add to message queue
	formattedMessage := fmt.Sprintf("File: %s\nLine: %s\nComment: %s\n\nPlease improve this code.", 
		message.File, message.Line, message.Comment)
	messageQueue = append(messageQueue, formattedMessage)

	// Clean up the message file
	os.Remove(messageFile)
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
			// Get latest message
			message := messageQueue[len(messageQueue)-1]
			result := map[string]interface{}{
				"contents": []map[string]interface{}{
					{
						"uri":      uri,
						"mimeType": "text/plain",
						"text":     message,
					},
				},
			}
			sendResponse(req.ID, result)
			// Clear the message after reading
			messageQueue = messageQueue[:len(messageQueue)-1]
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
			message := messageQueue[len(messageQueue)-1]
			messageQueue = messageQueue[:len(messageQueue)-1]
			
			result := map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": message,
					},
				},
			}
			sendResponse(req.ID, result)
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