#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
coverage_file="$repo_root/coverage.txt"
use_mise_exec=0
runner_label="PATH"

cleanup() {
  rm -f "$coverage_file"
}

fail_prereq() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

warn() {
  printf 'warning: %s\n' "$1" >&2
}

detect_runner() {
  if command -v mise >/dev/null 2>&1; then
    if (
      cd "$repo_root"
      mise exec -- true
    ) >/dev/null 2>&1; then
      use_mise_exec=1
      runner_label="mise"
      return
    fi

    warn "mise exec is unavailable in this environment; falling back to the current PATH toolchain. Ensure your shell is using the Go installed from mise.toml."
  fi

  if ! command -v go >/dev/null 2>&1; then
    fail_prereq "Go is required to run repo CI checks. Install the repo tools with 'mise install' and ensure the resulting toolchain is on PATH."
  fi
}

run_root() {
  (
    cd "$repo_root"
    if [[ "$use_mise_exec" -eq 1 ]]; then
      mise exec -- "$@"
    else
      "$@"
    fi
  )
}

run_cgotest() {
  (
    cd "$repo_root/cgotest"
    if [[ "$use_mise_exec" -eq 1 ]]; then
      CGO_ENABLED=1 mise exec -- "$@"
    else
      CGO_ENABLED=1 "$@"
    fi
  )
}

trap cleanup EXIT
detect_runner

printf '==> [%s] go build ./...\n' "$runner_label"
run_root go build ./...

printf '==> [%s] go vet ./...\n' "$runner_label"
run_root go vet ./...

printf '==> [%s] go test -v -race -coverprofile=coverage.txt ./...\n' "$runner_label"
run_root go test -v -race -coverprofile="$coverage_file" ./...

printf '==> (cd cgotest && CGO_ENABLED=1 go test -v -count=1 ./...)\n'
if ! run_cgotest go test -run '^$' ./... >/dev/null 2>&1; then
  fail_prereq "CGO oracle test prerequisites are incomplete. Ensure the H3 C library and linker paths are configured locally."
fi
run_cgotest go test -v -count=1 ./...
