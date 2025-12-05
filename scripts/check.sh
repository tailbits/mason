#!/usr/bin/env bash
set -euo pipefail

echo "Go version: $(go version)"

echo "==> Checking formatting"
fmt_out=$(gofmt -s -l .)
if [[ -n "${fmt_out}" ]]; then
  echo "The following files are not gofmt'ed:" >&2
  echo "${fmt_out}" >&2
  if [[ "${FIX_FORMAT:-}" == "1" ]]; then
    echo "Auto-fixing formatting (FIX_FORMAT=1)" >&2
    gofmt -s -w .
  else
    echo "Run: gofmt -s -w .  (or set FIX_FORMAT=1 to auto-fix)" >&2
    exit 1
  fi
fi

echo "==> go vet"
go vet ./...

echo "==> golangci-lint"
if command -v golangci-lint >/dev/null 2>&1; then
  golangci-lint run --timeout=5m
else
  echo "golangci-lint not found; skipping lint. Install: https://golangci-lint.run/welcome/install/" >&2
fi

echo "==> go test"
go test ./...

echo "All checks passed."

