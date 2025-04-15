#!/bin/bash

echo 'I want to use the file_edit tool to update test.txt: {"file_path": "test.txt", "operation": "replace", "content": "This is the new content", "start_line": 2, "end_line": 2}' | ./ai-agent
cat test.txt