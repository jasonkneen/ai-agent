package tools

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

// Tool defines the interface for agent tools
type Tool interface {
    Execute(input string) (string, error)
    GetName() string
    GetDescription() string
}

// WebSearchTool is a mock tool simulating a web search
type WebSearchTool struct{}

func (t *WebSearchTool) Execute(input string) (string, error) {
    // Mock a web search result
    query := strings.TrimSpace(input)
    return fmt.Sprintf("Mock search results for '%s': [Result 1, Result 2, Result 3]", query), nil
}

func (t *WebSearchTool) GetName() string {
    return "web_search"
}

func (t *WebSearchTool) GetDescription() string {
    return "Search the web for information"
}

// FileSearchTool is a tool for finding files in the system
type FileSearchTool struct {
    RootDir string
}

func (t *FileSearchTool) Execute(input string) (string, error) {
    searchTerm := strings.TrimSpace(input)
    var matchedFiles []string
    
    err := filepath.Walk(t.RootDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(searchTerm)) {
            matchedFiles = append(matchedFiles, path)
        }
        return nil
    })
    
    if err != nil {
        return "", err
    }
    
    if len(matchedFiles) == 0 {
        return fmt.Sprintf("No files matching '%s' found", searchTerm), nil
    }
    
    result := fmt.Sprintf("Files matching '%s':\n", searchTerm)
    for _, file := range matchedFiles {
        result += fmt.Sprintf("- %s\n", file)
    }
    return result, nil
}

func (t *FileSearchTool) GetName() string {
    return "file_search"
}

func (t *FileSearchTool) GetDescription() string {
    return "Search for files by name in the filesystem"
}

// FileReadTool reads the content of a file
type FileReadTool struct{}

func (t *FileReadTool) Execute(input string) (string, error) {
    filePath := strings.TrimSpace(input)
    
    // Check if file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return "", fmt.Errorf("file not found: %s", filePath)
    }
    
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    
    return string(data), nil
}

func (t *FileReadTool) GetName() string {
    return "file_read"
}

func (t *FileReadTool) GetDescription() string {
    return "Read the contents of a file at the specified path"
}

// FileEditRequest defines the structure for file edit operations
type FileEditRequest struct {
    FilePath  string `json:"file_path"`
    Operation string `json:"operation"` // "replace" or "append"
    Content   string `json:"content"`
    StartLine int    `json:"start_line,omitempty"` // Optional for replace operation
    EndLine   int    `json:"end_line,omitempty"`   // Optional for replace operation
}

// FileEditTool edits content in files
type FileEditTool struct{}

func (t *FileEditTool) Execute(input string) (string, error) {
    fmt.Printf("FileEditTool received input: %s\n", input)
    
    // Parse the JSON request
    var request FileEditRequest
    if err := json.Unmarshal([]byte(input), &request); err != nil {
        return "", fmt.Errorf("invalid JSON input: %w, input was: %s", err, input)
    }
    
    // Validate request
    if request.FilePath == "" {
        return "", fmt.Errorf("file_path is required")
    }
    if request.Operation == "" {
        return "", fmt.Errorf("operation is required")
    }
    
    // Check if file exists
    fileInfo, err := os.Stat(request.FilePath)
    if os.IsNotExist(err) {
        // If file doesn't exist and we're not appending, create it
        if request.Content != "" {
            if err := ioutil.WriteFile(request.FilePath, []byte(request.Content), 0644); err != nil {
                return "", fmt.Errorf("failed to create file: %w", err)
            }
            return fmt.Sprintf("Created new file %s with %d bytes", request.FilePath, len(request.Content)), nil
        }
        return "", fmt.Errorf("file not found: %s", request.FilePath)
    } else if err != nil {
        return "", fmt.Errorf("error accessing file: %w", err)
    }
    
    // Don't allow editing directories
    if fileInfo.IsDir() {
        return "", fmt.Errorf("%s is a directory, not a file", request.FilePath)
    }
    
    // Read the current file content
    currentContent, err := ioutil.ReadFile(request.FilePath)
    if err != nil {
        return "", fmt.Errorf("failed to read file: %w", err)
    }
    
    var newContent []byte
    
    switch request.Operation {
    case "replace":
        // Replace entire file
        if request.StartLine == 0 && request.EndLine == 0 {
            newContent = []byte(request.Content)
        } else {
            // Replace specific lines
            lines := strings.Split(string(currentContent), "\n")
            
            // Validate line numbers
            if request.StartLine < 1 || request.StartLine > len(lines) {
                return "", fmt.Errorf("start_line out of range: valid range is 1-%d", len(lines))
            }
            if request.EndLine < request.StartLine || request.EndLine > len(lines) {
                request.EndLine = len(lines)
            }
            
            // Replace the specified lines
            newLines := append(
                lines[:request.StartLine-1],
                append(
                    strings.Split(request.Content, "\n"),
                    lines[request.EndLine:]...,
                )...,
            )
            newContent = []byte(strings.Join(newLines, "\n"))
        }
    
    case "append":
        // Append to the file
        if strings.HasSuffix(string(currentContent), "\n") {
            newContent = append(currentContent, []byte(request.Content)...)
        } else {
            newContent = append(currentContent, []byte("\n"+request.Content)...)
        }
    
    default:
        return "", fmt.Errorf("unsupported operation: %s. Use 'replace' or 'append'", request.Operation)
    }
    
    // Write the modified content back to the file
    if err := ioutil.WriteFile(request.FilePath, newContent, fileInfo.Mode()); err != nil {
        return "", fmt.Errorf("failed to write to file: %w", err)
    }
    
    return fmt.Sprintf("Successfully edited %s (%d bytes written)", request.FilePath, len(newContent)), nil
}

func (t *FileEditTool) GetName() string {
    return "file_edit"
}

func (t *FileEditTool) GetDescription() string {
    return "Edit a file - can replace entire file, replace specific lines, or append content (input is JSON)"
}
