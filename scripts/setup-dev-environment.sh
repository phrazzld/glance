#!/bin/bash
# Glance Developer Environment Setup Script
# This script sets up the complete development environment for the Glance project.
# It installs all necessary tools, configures pre-commit hooks, and sets up the local environment.

set -e  # Exit on any error

# Text formatting
BOLD="\033[1m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
BLUE="\033[0;34m"
RED="\033[0;31m"
RESET="\033[0m"

echo -e "${BOLD}${BLUE}Glance Development Environment Setup${RESET}"
echo "This script will set up your development environment for the Glance project."
echo

# Function to check if command exists
command_exists() {
  command -v "$1" &> /dev/null
}

# Function to check OS
detect_os() {
  if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "linux"
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "macos"
  elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    echo "windows"
  else
    echo "unknown"
  fi
}

OS=$(detect_os)
echo -e "${BOLD}Detected OS:${RESET} ${OS}"
echo

# Step 1: Check Go version
echo -e "${BOLD}${BLUE}Checking Go installation...${RESET}"
if ! command_exists go; then
  echo -e "${RED}Go is not installed.${RESET} Please install Go 1.23 or higher before continuing."
  echo
  echo "Installation instructions:"
  echo "- macOS: brew install go"
  echo "- Linux: https://golang.org/doc/install"
  echo "- Windows: https://golang.org/doc/install"
  exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)

if [[ "$GO_MAJOR" -lt 1 || ("$GO_MAJOR" -eq 1 && "$GO_MINOR" -lt 23) ]]; then
  echo -e "${RED}Go version ${GO_VERSION} is too old.${RESET} Glance requires Go 1.23 or higher."
  echo "Please upgrade your Go installation."
  exit 1
fi

echo -e "${GREEN}✓ Go ${GO_VERSION} is installed.${RESET}"
echo

# Step 2: Check Git installation
echo -e "${BOLD}${BLUE}Checking Git installation...${RESET}"
if ! command_exists git; then
  echo -e "${RED}Git is not installed.${RESET} Please install Git before continuing."
  echo
  echo "Installation instructions:"
  echo "- macOS: brew install git"
  echo "- Linux: sudo apt-get install git or sudo yum install git"
  echo "- Windows: https://git-scm.com/download/win"
  exit 1
fi

GIT_VERSION=$(git --version | awk '{print $3}')
echo -e "${GREEN}✓ Git ${GIT_VERSION} is installed.${RESET}"
echo

# Step 3: Configure Git settings
echo -e "${BOLD}${BLUE}Checking Git configuration...${RESET}"
if [[ -z "$(git config --global user.name)" ]]; then
  echo -e "${YELLOW}Git user.name is not configured.${RESET}"
  read -p "Enter your name for Git commits: " GIT_USER_NAME
  git config --global user.name "$GIT_USER_NAME"
else
  echo -e "${GREEN}✓ Git user.name is configured as: $(git config --global user.name)${RESET}"
fi

if [[ -z "$(git config --global user.email)" ]]; then
  echo -e "${YELLOW}Git user.email is not configured.${RESET}"
  read -p "Enter your email for Git commits: " GIT_USER_EMAIL
  git config --global user.email "$GIT_USER_EMAIL"
else
  echo -e "${GREEN}✓ Git user.email is configured as: $(git config --global user.email)${RESET}"
fi
echo

# Step 4: Set up pre-commit hooks
echo -e "${BOLD}${BLUE}Setting up pre-commit hooks...${RESET}"
# Using the existing setup-precommit.sh script
if [[ -f "$(dirname "$0")/setup-precommit.sh" ]]; then
  echo "Running pre-commit setup script..."
  bash "$(dirname "$0")/setup-precommit.sh"
else
  echo -e "${RED}Error: setup-precommit.sh script not found.${RESET}"
  echo "This script should be in the same directory as setup-dev-environment.sh."
  exit 1
fi
echo

# Step 5: Check and install GitHub CLI (optional)
echo -e "${BOLD}${BLUE}Checking GitHub CLI installation (optional)...${RESET}"
if ! command_exists gh; then
  echo -e "${YELLOW}GitHub CLI (gh) is not installed.${RESET}"
  echo "The GitHub CLI is recommended for working with GitHub Actions, PRs, and issues."
  echo

  read -p "Would you like to install the GitHub CLI? (y/n): " INSTALL_GH
  if [[ "$INSTALL_GH" == "y" || "$INSTALL_GH" == "Y" ]]; then
    if [[ "$OS" == "macos" ]]; then
      echo "Installing GitHub CLI via Homebrew..."
      brew install gh
    elif [[ "$OS" == "linux" ]]; then
      echo "Please follow the installation instructions at: https://github.com/cli/cli#installation"
      echo "For Ubuntu/Debian:"
      echo "  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg"
      echo "  echo \"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null"
      echo "  sudo apt update"
      echo "  sudo apt install gh"
    elif [[ "$OS" == "windows" ]]; then
      echo "Please follow the installation instructions at: https://github.com/cli/cli#installation"
      echo "For Windows:"
      echo "  winget install --id GitHub.cli"
      echo "  or"
      echo "  scoop install gh"
    fi
  else
    echo "Skipping GitHub CLI installation."
  fi
else
  GH_VERSION=$(gh --version | head -1 | awk '{print $3}')
  echo -e "${GREEN}✓ GitHub CLI ${GH_VERSION} is installed.${RESET}"
fi
echo

# Step 6: Set up local environment
echo -e "${BOLD}${BLUE}Setting up local environment...${RESET}"

# Check if .env file exists, if not create it from template
if [[ ! -f ".env" ]]; then
  echo "Creating .env file template..."
  cat > .env.example << EOF
# Glance environment configuration

# Google Generative AI API key (required)
GEMINI_API_KEY=your_api_key_here

# Optional configuration:
# LOG_LEVEL=debug  # Logging level (debug, info, warn, error)
EOF

  cp .env.example .env
  echo -e "${YELLOW}A new .env file has been created.${RESET}"
  echo "You need to edit it and add your GEMINI_API_KEY."
else
  echo -e "${GREEN}✓ .env file already exists.${RESET}"

  # Check if GEMINI_API_KEY is set
  if grep -q "GEMINI_API_KEY=your_api_key_here" .env || grep -q "GEMINI_API_KEY=" .env; then
    echo -e "${YELLOW}Warning: GEMINI_API_KEY appears to be unset or using the default value.${RESET}"
    echo "You need to edit the .env file and add your actual API key for Glance to work."
  fi
fi
echo

# Step 7: Verify Go modules
echo -e "${BOLD}${BLUE}Verifying Go modules...${RESET}"
echo "Running go mod verify..."
go mod verify
echo "Running go mod tidy..."
go mod tidy
echo -e "${GREEN}✓ Go modules verified and tidied.${RESET}"
echo

# Step 8: Try a test build
echo -e "${BOLD}${BLUE}Building the project...${RESET}"
echo "Running go build..."
go build -o glance
echo -e "${GREEN}✓ Build successful.${RESET}"
echo

# Step 9: Display final instructions
echo -e "${BOLD}${GREEN}Setup Complete!${RESET}"
echo
echo -e "${BOLD}Next steps:${RESET}"
echo
echo "1. Ensure your GEMINI_API_KEY is properly set in the .env file"
echo "2. Run tests with: go test ./..."
echo "3. Run the application with: ./glance [--verbose] /path/to/directory"
echo
echo "For more information:"
echo "- README.md: Overview and basic usage"
echo "- docs/PRECOMMIT.md: Pre-commit hook details"
echo "- docs/GITHUB_ACTIONS.md: CI/CD workflow details"
echo
echo -e "${BOLD}${BLUE}Happy coding!${RESET}"
