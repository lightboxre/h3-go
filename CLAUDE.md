# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`h3-native-go` — a native Go implementation of the H3 geospatial indexing system.

> This repository is currently empty. Update this file as the project takes shape.

# H3 Native Go — Agent Orchestration Rules

## Execution Model
This project uses a phased parallel implementation plan defined in `implementation-plan.md`.
Always read `implementation-plan.md` before beginning any work.

## Subagent Delegation Rules
When executing a phase that contains multiple independent agents (Phases 2, 4, 5, 6),
ALWAYS spawn parallel subagents — never execute them sequentially in the same session.

Spawn a subagent for each labeled agent (Agent A, Agent B, etc.) in the current phase.
Each subagent must receive:
1. Its specific agent label and scope from the plan
2. The full list of files it must create
3. Its dependencies (which prior phase outputs it reads)
4. The correctness constraints section from the plan

## Phase Sequencing
- Wait for ALL subagents in a phase to complete before starting the next phase.
- Phases 1 and 3 are sequential (single agent) — do not spawn subagents.
- Phases 2, 4, 5, 6 are parallel — always spawn subagents.

## Domain Ownership (no cross-agent file edits)
- Agent A → internal/h3index/
- Agent B → internal/coordijk/
- Agent C → internal/bbox/
- Agent D → h3.go (public API)
- Agent E → internal/algos/
- Agent F → internal/polygon/
- Agent G → h3.go (directed edge functions only)
- Agent H → h3.go (vertex functions only)
- Agent I → internal/math/
- Agent J → internal/testutil/, testdata/, h3_test.go
- Agent K → h3_cgo_test.go, h3_bench_test.go, BENCHMARKS.md

## Common Commands

Once a `go.mod` is initialized:

```bash
# Build
go build ./...

# Test
go test ./...

# Run a single test
go test ./... -run TestName

# Lint (if golangci-lint is configured)
golangci-lint run
```
