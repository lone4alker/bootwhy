#!/usr/bin/env bash
# ==============================================================================
# BootWhy Entrypoint Runner
# Single-file runner for analyzing, explaining, testing, and installing BootWhy
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# ANSI Colors
BOLD="\033[1m"
GREEN="\033[32m"
BLUE="\033[34m"
YELLOW="\033[33m"
RESET="\033[0m"

# Ensure Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo "Error: 'go' is not installed or not in PATH." >&2
    exit 1
fi

# Ensure binary is built and up to date
build_binary() {
    if [ ! -f "bootwhy" ] || [ "cmd/bootwhy/main.go" -nt "bootwhy" ]; then
        echo -e "${BLUE}==>${RESET} Building bootwhy..."
        go build -o bootwhy ./cmd/bootwhy
    fi
}

show_help() {
    echo -e "${BOLD}BootWhy Runner${RESET}"
    echo
    echo "Usage:"
    echo "  ./run.sh                 Run live systemd boot analysis on this machine"
    echo "  ./run.sh explain <row>   Explain root cause and remediation for finding row (e.g. ./run.sh explain 1)"
    echo "  ./run.sh --json          Export full diagnostic analysis as JSON"
    echo "  ./run.sh --from <dir>    Analyze an offline fixture (e.g. ./run.sh --from testdata/faults/device-timeout)"
    echo "  ./run.sh test            Run the automated test suite and accuracy checks"
    echo "  ./run.sh install         Install 'bootwhy' to ~/.local/bin so you can run 'bootwhy' from anywhere"
    echo "  ./run.sh help            Show this help message"
    echo
}

# Subcommands
case "${1:-}" in
    help|--help|-h)
        show_help
        exit 0
        ;;
    test)
        echo -e "${BLUE}==>${RESET} Running unit and accuracy tests across all fixtures..."
        go test -v ./...
        exit 0
        ;;
    install)
        build_binary
        mkdir -p "$HOME/.local/bin"
        install -m 755 bootwhy "$HOME/.local/bin/bootwhy"
        echo -e "${GREEN}✓ Installed successfully to $HOME/.local/bin/bootwhy${RESET}"
        echo "You can now run 'bootwhy' directly from any terminal window."
        exit 0
        ;;
    *)
        build_binary
        ./bootwhy "$@"
        ;;
esac
