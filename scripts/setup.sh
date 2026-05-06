#!/bin/bash
# Obsidian Notes Setup Script
# 一键安装脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Obsidian Notes Setup ==="
echo "Project directory: $PROJECT_DIR"
cd "$PROJECT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check Python version
echo -e "\n${YELLOW}Checking Python version...${NC}"
PYTHON_CMD=""
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
echo -e "\n${YELLOW}Creating virtual environment...${NC}"
if [[ ! -d "venv" ]]; then
    $PYTHON_CMD -m venv venv
    echo -e "${GREEN}Virtual environment created${NC}"
else
    echo -e "${GREEN}Virtual environment already exists${NC}"
fi

# Activate virtual environment
source venv/bin/activate

# Upgrade pip
echo -e "\n${YELLOW}Upgrading pip...${NC}"
pip install --upgrade pip

# Install dependencies
echo -e "\n${YELLOW}Installing dependencies...${NC}"
pip install pyyaml

# Create required directories
echo -e "\n${YELLOW}Creating directories...${NC}"
mkdir -p config plugins scripts outputs raw wiki

# Create default config if not exists
echo -e "\n${YELLOW}Setting up configuration...${NC}"
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/obsidian"
mkdir -p "$CONFIG_DIR"
if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
    if [[ -f "config/default_config.yaml" ]]; then
        cp config/default_config.yaml "$CONFIG_DIR/config.yaml"
        echo -e "${GREEN}Default config copied to $CONFIG_DIR/config.yaml${NC}"
    fi
fi

# Create sample plugin
echo -e "\n${YELLOW}Creating sample plugin...${NC}"
if [[ ! -f "plugins/example.py" ]]; then
    cat > plugins/example.py << 'EOF'
"""Example plugin."""

from plugins import Plugin


class ExamplePlugin(Plugin):
    """Example plugin demonstrating the plugin system."""
    
    name = "example"
    version = "1.0.0"

    def on_load(self):
        print(f"Example plugin loaded (v{self.version})")

    def on_enable(self):
        print("Example plugin enabled")

    def on_disable(self):
        print("Example plugin disabled")


# Plugin instance
plugin = ExamplePlugin()
EOF
    echo -e "${GREEN}Sample plugin created${NC}"
fi

# Set permissions
chmod +x scripts/*.sh 2>/dev/null || true

echo -e "\n${GREEN}=== Setup Complete ===${NC}"
echo ""
echo "To activate the virtual environment:"
echo "  source venv/bin/activate"
echo ""
echo "To run the application:"
echo "  python -m obsidian"
echo ""
echo "Config file: $CONFIG_DIR/config.yaml"
