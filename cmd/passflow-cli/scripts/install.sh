#!/bin/bash
set -e

# Passflow CLI Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/jaak-ai/passflow-cli/main/scripts/install.sh | bash

REPO="jaak-ai/passflow-cli"
BINARY_NAME="passflow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*)  echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest version from GitHub
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
install() {
    local os=$(detect_os)
    local arch=$(detect_arch)
    local version="${VERSION:-$(get_latest_version)}"

    if [ -z "$version" ]; then
        error "Could not determine latest version. Please set VERSION env var."
    fi

    info "Installing Passflow CLI ${version} for ${os}/${arch}..."

    local filename="${BINARY_NAME}_${version#v}_${os}_${arch}"
    if [ "$os" = "windows" ]; then
        filename="${filename}.zip"
    else
        filename="${filename}.tar.gz"
    fi

    local url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    local tmp_dir=$(mktemp -d)

    info "Downloading from ${url}..."

    if ! curl -fsSL "$url" -o "${tmp_dir}/${filename}"; then
        error "Failed to download ${url}"
    fi

    info "Extracting..."
    if [ "$os" = "windows" ]; then
        unzip -q "${tmp_dir}/${filename}" -d "${tmp_dir}"
    else
        tar -xzf "${tmp_dir}/${filename}" -C "${tmp_dir}"
    fi

    info "Installing to ${INSTALL_DIR}..."

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        warn "Need sudo to install to ${INSTALL_DIR}"
        sudo mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    rm -rf "$tmp_dir"

    info "Passflow CLI installed successfully!"
    echo ""
    echo "Run 'passflow --help' to get started."
    echo ""
    echo "Quick setup:"
    echo "  passflow config set api-url https://api.passflow.ai"
    echo "  passflow login --token <your-jwt-token>"
}

install
