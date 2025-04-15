package llm

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "github.com/joho/godotenv"
)

// Message mirrors agent.Message for LLM requests
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// Client manages Anthropic Claude API interactions
type Client struct {
    apiKey   string
    endpoint string
}

// NewClient initializes an Anthropic Claude client
func NewClient() *Client {
    // Load .env for API key
    _ = godotenv.Load()
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    // Default to Anthropic API endpoint
    endpoint := os.Getenv("ANTHROPIC_ENDPOINT")
    if endpoint == "" {
        endpoint = "https://api.anthropic.com/v1/messages"
    }
    return &Client{
        apiKey:   apiKey,
        endpoint: endpoint,
    }
}

// Query sends a request to Claude and returns the response
func (c *Client) Query(messages []Message) (string, error) {
    if c.apiKey == "" {
        // Mock response if no API key
        return fmt.Sprintf("Mock Claude response to: %s", messages[len(messages)-1].Content), nil
    }

    // Extract system message if present
    var systemPrompt string
    var apiMessages []map[string]string
    
    for _, msg := range messages {
        if msg.Role == "system" {
            systemPrompt = msg.Content
        } else if msg.Role == "user" || msg.Role == "assistant" {
            // Keep original role for user and assistant
            apiMessages = append(apiMessages, map[string]string{
                "role":    msg.Role,
                "content": msg.Content,
            })
        } else if msg.Role == "tool" {
            // Convert tool messages to user messages for API compatibility
            apiMessages = append(apiMessages, map[string]string{
                "role":    "user",
                "content": fmt.Sprintf("[Tool Output] %s", msg.Content),
            })
        }
    }

    // Prepare request payload
    payload := map[string]interface{}{
        "model":      "claude-3-5-sonnet-20241022", // Use Claude 3.5 Sonnet
        "max_tokens": 1024,
        "messages":   apiMessages,
    }
    
    // Add system message if present
    if systemPrompt != "" {
        payload["system"] = systemPrompt
    }
    body, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("failed to marshal payload: %v", err)
    }

    // Create HTTP request
    req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(body))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", c.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    // Also set Authorization header as Bearer token
    req.Header.Set("Authorization", "Bearer "+c.apiKey)

    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to send request: %v", err)
    }
    defer resp.Body.Close()
    
    // Check response status
    if resp.StatusCode != http.StatusOK {
        var errorResponse map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
            return "", fmt.Errorf("API error: status %d", resp.StatusCode)
        }
        return "", fmt.Errorf("API error: status %d, message: %v", resp.StatusCode, errorResponse)
    }

    // Parse response
    var result struct {
        Content []struct {
            Type string `json:"type"`
            Text string `json:"text"`
        } `json:"content"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to decode response: %v", err)
    }
    if len(result.Content) == 0 {
        return "", fmt.Errorf("no response from Claude")
    }
    
    // Combine all text blocks
    var responseText string
    for _, content := range result.Content {
        if content.Type == "text" {
            responseText += content.Text
        }
    }
    
    if responseText == "" {
        return "", fmt.Errorf("no text content in response")
    }
    
    return responseText, nil
}
