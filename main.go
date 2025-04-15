package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    
    "jkneen.ai-agent/agent"
)

func main() {
    // Initialize the agent with a context file
    ag, err := agent.NewAgent("conversation.json")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize agent: %v\n", err)
        os.Exit(1)
    }
    defer ag.SaveContext() // Save context on exit

    fmt.Println("Welcome to the AI Agent (powered by Claude)! Type 'exit' to quitPo.")
    scanner := bufio.NewScanner(os.Stdin)

    for {
        fmt.Print("> ")
        if !scanner.Scan() {
            break
        }
        input := strings.TrimSpace(scanner.Text())
        if input == "exit" {
            break
        }
        if input == "" {
            continue
        }

        // Process the input
        response, err := ag.Process(input)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            continue
        }
        fmt.Println(response)
    }

    if err := scanner.Err(); err != nil {
        fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
    }
}
