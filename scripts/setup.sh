#!/bin/bash
# Battlegrounds Advisor - Setup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================"
echo "Battlegrounds Advisor - Setup"
echo "========================================"

# Check Python
echo -e "\n${YELLOW}[Step 1/3] Checking Python...${NC}"
for cmd in python3 python; do
    if command -v $cmd &> /dev/null; then
        version=$($cmd --version 2>&1 | awk '{print $2}')
        major=$(echo $version | cut -d. -f1)
        minor=$(echo $version | cut -d. -f2)
        if [[ $major -eq 3 && $minor -ge 8 ]] || [[ $major -gt 3 ]]; then
            PYTHON_CMD=$cmd
            echo -e "${GREEN}Found Python $version${NC}"
            break
        fi
    fi
done

if [[ -z "$PYTHON_CMD" ]]; then
    echo -e "${RED}Error: Python 3.8+ required${NC}"
    exit 1
fi

# Create virtual environment
echo -e "\n${YELLOW}[Step 2/3] Creating virtual environment...${NC}"
if [[ ! -d ".venv" ]]; then
    $PYTHON_CMD -m venv .venv
    echo -e "${GREEN}Virtual environment created${NC}"
else
    echo -e "${GREEN}Virtual environment already exists${NC}"
fi

# Activate and install dependencies
source .venv/bin/activate

echo -e "\n${YELLOW}[Step 3/3] Installing dependencies...${NC}"
pip install --upgrade pip
pip install -r reader/requirements.txt
echo -e "${GREEN}Dependencies installed${NC}"

# Run mock tests
echo -e "\n${YELLOW}Running Mock Tests...${NC}"
python tests/test_mock.py

echo -e "\n${GREEN}========================================"
echo "Setup Complete!"
echo "========================================${NC}"
