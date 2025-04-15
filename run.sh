#!/bin/bash

# Navigate to project directory
cd "$(dirname "$0")" || exit 1

# Ensure Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go: https://go.dev/doc/install"
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
go mod tidy
go get github.com/joho/godotenv

# Build the program
echo "Building the agent..."
go build -o ai-agent || {
    echo "Build failed"
    exit 1
}

# Run the program
echo "Running the agent..."
./ai-agent
