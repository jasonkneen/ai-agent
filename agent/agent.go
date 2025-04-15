package agent

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"
    
    "jkneen.ai-agent/llm"
    "jkneen.ai-agent/tools"
)

// Message represents a single message in the conversation
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// Agent holds the state and logic for the AI agent
type Agent struct {
    context      []Message
    contextFile  string
    llmClient    *llm.Client
    toolRegistry map[string]tools.Tool
}

// NewAgent initializes a new agent with a context file
func NewAgent(contextFile string) (*Agent, error) {
    llmClient := llm.NewClient() // Initialize Claude client
    
    // Create system message with tool descriptions
    systemMessage := "You are a helpful AI assistant powered by Claude. You have access to these tools:\n\n"
    
    // Will be populated with registered tools
    toolRegistry := make(map[string]tools.Tool)
    
    // Register all available tools
    webSearchTool := &tools.WebSearchTool{}
    fileSearchTool := &tools.FileSearchTool{RootDir: "."}
    fileReadTool := &tools.FileReadTool{}
    fileEditTool := &tools.FileEditTool{}
    
    // Add tools to registry
    toolRegistry[webSearchTool.GetName()] = webSearchTool
    toolRegistry[fileSearchTool.GetName()] = fileSearchTool
    toolRegistry[fileReadTool.GetName()] = fileReadTool
    toolRegistry[fileEditTool.GetName()] = fileEditTool
    
    // Build system message with tool descriptions
    for _, tool := range toolRegistry {
        systemMessage += fmt.Sprintf("- %s: %s\n", tool.GetName(), tool.GetDescription())
    }
    
    systemMessage += "\nTo use a tool, simply mention its name and what you want to do with it. For example: 'I need to use the web_search tool to find information about...' or 'I'll use file_search to look for...'."
    
    // Add special instructions for the file_edit tool
    systemMessage += "\n\nTo use the file_edit tool, include a JSON object with the following structure:"
    systemMessage += "\n```json"
    systemMessage += "\n{"
    systemMessage += "\n  \"file_path\": \"path/to/file.txt\", // Required: Path to the file to edit"
    systemMessage += "\n  \"operation\": \"replace\", // Required: Either 'replace' or 'append'"
    systemMessage += "\n  \"content\": \"new content\", // Required: The content to write"
    systemMessage += "\n  \"start_line\": 1, // Optional: Line number to start replacing (only for replace)"
    systemMessage += "\n  \"end_line\": 5 // Optional: Line number to end replacing (only for replace)"
    systemMessage += "\n}"
    systemMessage += "\n```"
    systemMessage += "\nFor example: 'I'll use the file_edit tool to update the README.md file: {\"file_path\": \"README.md\", \"operation\": \"replace\", \"content\": \"# Updated README\"}'."
    
    ag := &Agent{
        context:      []Message{{Role: "system", Content: systemMessage}},
        contextFile:  contextFile,
        llmClient:    llmClient,
        toolRegistry: toolRegistry,
    }

    // Load existing context if available
    if err := ag.loadContext(); err != nil {
        // Ignore if file doesn't exist
        if !os.IsNotExist(err) {
            return nil, err
        }
    }

    return ag, nil
}

// loadContext reads the conversation context from a file
func (a *Agent) loadContext() error {
    data, err := os.ReadFile(a.contextFile)
    if err != nil {
        return err
    }
    return json.Unmarshal(data, &a.context)
}

// SaveContext writes the conversation context to a file
func (a *Agent) SaveContext() error {
    data, err := json.MarshalIndent(a.context, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(a.contextFile, data, 0644)
}

// Process handles user input and returns a response
func (a *Agent) Process(input string) (string, error) {
    // Add user message to context
    a.context = append(a.context, Message{Role: "user", Content: input})

    // First get a response from the LLM
    llmMessages := convertToLLMMessages(a.context)
    llmResponse, err := a.llmClient.Query(llmMessages)
    if err != nil {
        return "", err
    }
    
    // Check if the LLM response wants to use a tool
    if a.shouldUseTool(llmResponse) {
        // Detect which tool to use from the response
        toolName := a.detectToolName(llmResponse)
        
        // Add the LLM's "I want to use a tool" response to the context
        a.context = append(a.context, Message{Role: "assistant", Content: llmResponse})
        
        // Execute the tool
        toolResponse, err := a.executeTool(toolName, llmResponse)
        if err != nil {
            return "", fmt.Errorf("tool execution error: %w", err)
        }
        
        // Add tool response to context
        toolRoleMessage := fmt.Sprintf("Tool '%s' returned: %s", toolName, toolResponse)
        a.context = append(a.context, Message{Role: "tool", Content: toolRoleMessage})
        
        // Get final response from LLM with tool results
        llmMessages = convertToLLMMessages(a.context)
        finalResponse, err := a.llmClient.Query(llmMessages)
        if err != nil {
            return "", err
        }
        
        // Add final response to context
        a.context = append(a.context, Message{Role: "assistant", Content: finalResponse})
        return finalResponse, nil
    }

    // No tool needed, just return the LLM response
    a.context = append(a.context, Message{Role: "assistant", Content: llmResponse})
    return llmResponse, nil
}

// shouldUseTool determines if a tool is needed based on LLM response
func (a *Agent) shouldUseTool(input string) bool {
    for toolName := range a.toolRegistry {
        if strings.Contains(strings.ToLower(input), strings.ToLower(toolName)) {
            return true
        }
    }
    // Also check for explicit tool usage phrases
    return strings.Contains(strings.ToLower(input), "use the") && 
           strings.Contains(strings.ToLower(input), "tool")
}

// detectToolName extracts the tool name from LLM response
func (a *Agent) detectToolName(input string) string {
    for toolName := range a.toolRegistry {
        if strings.Contains(strings.ToLower(input), strings.ToLower(toolName)) {
            return toolName
        }
    }
    return "web_search" // fallback to web_search if no specific tool detected
}

// extractToolInput extracts the input for the tool from LLM response
func (a *Agent) extractToolInput(toolName, input string) string {
    // For file_edit tool, extract JSON content
    if toolName == "file_edit" {
        // Find JSON in the input
        jsonStart := strings.Index(input, "{")
        jsonEnd := strings.LastIndex(input, "}")
        
        if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
            return input[jsonStart : jsonEnd+1]
        }
        
        // If we couldn't extract JSON, return a helpful error message
        return `{"error": "Could not extract valid JSON from input. Please provide a valid JSON object with file_path, operation, and content fields."}`
    }
    
    // Regular extraction for other tools
    toolNameIndex := strings.Index(strings.ToLower(input), strings.ToLower(toolName))
    if toolNameIndex == -1 {
        return input
    }
    
    // Get content after tool name
    afterToolName := input[toolNameIndex+len(toolName):]
    
    // Clean up - remove common phrases that might appear
    phrases := []string{"tool", "to", "with", "for", "using", "use", "the", ":"}
    cleanInput := afterToolName
    for _, phrase := range phrases {
        cleanInput = strings.ReplaceAll(cleanInput, phrase, "")
    }
    
    return strings.TrimSpace(cleanInput)
}

// executeTool runs a registered tool
func (a *Agent) executeTool(toolName, input string) (string, error) {
    tool, exists := a.toolRegistry[toolName]
    if !exists {
        return "", fmt.Errorf("tool %s not found", toolName)
    }
    
    // Extract actual input for the tool
    toolInput := a.extractToolInput(toolName, input)
    
    return tool.Execute(toolInput)
}

// convertToLLMMessages converts agent messages to LLM messages
func convertToLLMMessages(agentMessages []Message) []llm.Message {
    llmMessages := make([]llm.Message, len(agentMessages))
    for i, msg := range agentMessages {
        llmMessages[i] = llm.Message{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }
    return llmMessages
}
