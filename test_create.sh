#!/bin/bash

# Remove the file if it exists
rm -f new_file.txt

echo 'I want to use the file_edit tool to create a new file: {"file_path": "new_file.txt", "operation": "replace", "content": "This is a newly created file\nIt was created using the file_edit tool\nWith multiple lines"}' | ./ai-agent
cat new_file.txt