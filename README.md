# Go AI Agent

A Go implementation of an AI agent capable of using tools and maintaining conversation state.

## Overview

This project implements an AI agent in Go that can:
- Process natural language inputs
- Use various tools to accomplish tasks
- Maintain conversation context
- Generate appropriate responses

## Getting Started

### Prerequisites

- Go 1.21 or later

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/ai-agent.git
cd ai-agent

# Install dependencies
go mod tidy

# Build the binary
go build -o ai-agent
```

### Usage

```bash
# Run the agent
./ai-agent
```

Or use the provided script:

```bash
./run.sh
```

## Project Structure

- `agent/`: Contains the core agent implementation
- `llm/`: LLM client implementation
- `tools/`: Definition of tools the agent can use
- `main.go`: Entry point for the application

## Credits

This project is based on the guide [How to Build an Agent](https://ampcode.com/how-to-build-an-agent) by Thorsten Ball.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please check out our [Contributing Guide](CONTRIBUTING.md) for details on how to get started.