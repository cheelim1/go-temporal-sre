#!/bin/bash
# Demo 1: Setup - Build the application only
set -e

echo -e "\033[1;32m=== Demo 1: Building SuperScript ===\033[0m"
echo -e "\033[1;33mBuilding the SuperScript application\033[0m"
cd /Users/leow/GOMOD/go-temporal-sre
go build -o bin/superscript ./cmd/superscript/

echo -e "\033[1;32mBuild complete! The binary is now ready.\033[0m"
echo -e "\033[1;33mNext steps:\033[0m"
echo -e "  1. Start Temporal server with: make start-temporal"
echo -e "  2. Start SuperScript with: make superscript-start"
echo -e "  3. Run demos with: make superscript-demo-2, demo-3, etc."
echo -e "  4. When done, stop with: make superscript-stop"
