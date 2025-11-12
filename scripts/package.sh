#!/bin/bash
# Package MTGA Companion for distribution
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}==> Packaging MTGA Companion${NC}"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     PLATFORM=linux;;
    Darwin*)    PLATFORM=darwin;;
    MINGW*|MSYS*|CYGWIN*) PLATFORM=windows;;
    *)          PLATFORM="unknown";;
esac

echo -e "${GREEN}==> Detected platform: ${PLATFORM}${NC}"

# Check if fyne command is available
if ! command -v fyne &> /dev/null; then
    echo -e "${YELLOW}fyne command not found. Installing...${NC}"
    go install fyne.io/fyne/v2/cmd/fyne@latest
fi

# Check if Icon.png exists
if [ ! -f "Icon.png" ]; then
    echo -e "${YELLOW}Warning: Icon.png not found. Using default icon.${NC}"
    echo -e "${YELLOW}Create Icon.png (512x512) for custom application icon.${NC}"
fi

# Clean previous builds
echo -e "${GREEN}==> Cleaning previous builds${NC}"
rm -rf dist/
mkdir -p dist/

# Build executable first
echo -e "${GREEN}==> Building executable${NC}"
case "${PLATFORM}" in
    linux)
        go build -v -o mtga-companion ./cmd/mtga-companion
        ;;
    darwin)
        go build -v -o mtga-companion ./cmd/mtga-companion
        ;;
    windows)
        go build -v -ldflags="-H windowsgui" -o mtga-companion.exe ./cmd/mtga-companion
        ;;
esac

# Package application
echo -e "${GREEN}==> Packaging application${NC}"
case "${PLATFORM}" in
    linux)
        if [ -f "Icon.png" ]; then
            fyne package -os linux -icon Icon.png -name "MTGA Companion" -appID com.github.rdhamilton.mtga-companion
        else
            fyne package -os linux -name "MTGA Companion" -appID com.github.rdhamilton.mtga-companion
        fi

        # Create tar.gz
        tar -czf dist/mtga-companion-linux-amd64.tar.gz mtga-companion README.md LICENSE
        echo -e "${GREEN}✓ Created dist/mtga-companion-linux-amd64.tar.gz${NC}"
        ;;

    darwin)
        if [ -f "Icon.png" ]; then
            fyne package -os darwin -icon Icon.png -name "MTGA Companion" -appID com.github.rdhamilton.mtga-companion
        else
            fyne package -os darwin -name "MTGA Companion" -appID com.github.rdhamilton.mtga-companion
        fi

        # Create tar.gz with .app bundle
        tar -czf dist/mtga-companion-macos-universal.tar.gz "MTGA Companion.app" README.md
        echo -e "${GREEN}✓ Created dist/mtga-companion-macos-universal.tar.gz${NC}"
        ;;

    windows)
        if [ -f "Icon.png" ]; then
            fyne package -os windows -icon Icon.png -name "MTGA-Companion" -appID com.github.rdhamilton.mtga-companion
        else
            fyne package -os windows -name "MTGA-Companion" -appID com.github.rdhamilton.mtga-companion
        fi

        # Create zip
        zip -r dist/mtga-companion-windows-amd64.zip mtga-companion.exe README.md LICENSE
        echo -e "${GREEN}✓ Created dist/mtga-companion-windows-amd64.zip${NC}"
        ;;

    *)
        echo -e "${RED}Error: Unsupported platform: ${PLATFORM}${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}==> Packaging complete!${NC}"
echo -e "Distribution files created in: ${GREEN}dist/${NC}"
ls -lh dist/
