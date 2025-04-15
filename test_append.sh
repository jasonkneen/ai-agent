#!/bin/bash

echo 'I want to use the file_edit tool to append to test.txt: {"file_path": "test.txt", "operation": "append", "content": "This is a new line appended to the file"}' | ./ai-agent
cat test.txt