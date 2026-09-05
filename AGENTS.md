# AGENTS.md

## Project overview

This repository contains a Go CLI tool for finding and cleaning stale dependency folders such as `node_modules`, `vendor`, `.venv`, and `target` directories.

Primary docs:

- [README.md](README.md)
- [docs/prd.md](docs/prd.md)
- [docs/technical-spec.draft.md](docs/technical-spec.draft.md)

## Working conventions

- Before starting any task or plan, read the most recent relevant artifacts first: [docs/handoffs](docs/handoffs/), [docs/prd.md](docs/prd.md), and any `*.spec*.md` or draft spec files under [docs/](docs/).
- For handoffs and specs, prefer the newest file first by modified date or recency, then read the relevant wider context only if needed.
- Do not start a task by reading the entire repo broadly; anchor the work in the newest handoff/spec and then widen only to the specific files required.
- When a task involves a prior session or follow-up work, prefer scanning [docs/handoffs](docs/handoffs/) and the spec docs before reading implementation files.
- Prefer small, focused changes that match the current package boundaries.
- Keep destructive behavior behind explicit user confirmation or a dry-run path.
- Treat filesystem operations as safety-critical: validate paths, skip inaccessible directories, and avoid broad recursive deletes.
- Preserve the existing CLI-first design; do not introduce a GUI or unrelated subsystem unless the task explicitly calls for it.
- Prefer updating existing package APIs and tests over creating new abstraction layers unnecessarily.

## Required artifact order for new work

1. Read the latest handoff under [docs/handoffs](docs/handoffs/).
2. Read the current product/spec documents, especially [docs/prd.md](docs/prd.md) and any relevant `*.spec*.md` or draft spec files under [docs/](docs/).
3. If the task is broader than a single module, keep the spec-first pattern: read the newest handoff/spec, then only widen to the relevant implementation files in [cmd/](cmd/), [internal/](internal/), and [pkg/](pkg/).
4. Only after that should the agent read the relevant implementation files in [cmd/](cmd/), [internal/](internal/), and [pkg/](pkg/).
5. If the task is broader than a single module, widen scope in small, purposeful increments rather than broad repo exploration.

## Repository layout

- [cmd/](cmd/) — Cobra command entry points and CLI wiring
- [internal/analyzer/](internal/analyzer/) — size and metadata analysis
- [internal/cache/](internal/cache/) — cache persistence and invalidation
- [internal/cleaner/](internal/cleaner/) — deletion logic and safety checks
- [internal/config/](internal/config/) — Viper config and defaults
- [internal/logger/](internal/logger/) — structured logging
- [internal/scanner/](internal/scanner/) — folder traversal and detection
- [internal/ui/](internal/ui/) — terminal formatting and selection UI
- [pkg/models/](pkg/models/) — shared domain types
- [pkg/utils/](pkg/utils/) — filesystem helper logic and target-folder matching
- [scripts/](scripts/) — build and release support

## Build and validation

Run commands from the repository root:

```bash
make test
make build
make fmt
```

Common Go checks:

```bash
go test ./...
go fmt ./...
```

## Safety expectations

- Treat cleanup as destructive by default; align with the PRD’s preview/confirm workflow.
- Do not make irreversible file deletions in dry-run or test paths.
- Prefer graceful skip-and-log behavior for permission errors, symlink edge cases, and unreadable directories.
- Keep config and runtime state under the app’s managed directory (for example, `$HOME/.depocleaner`), not in project folders.

## Common pitfalls

- Do not bypass the package structure for simple changes; keep scanner, analyzer, cleaner, and UI responsibilities separated.
- The cache should remain a performance optimization, not a source of incorrect scan results.
- If a change affects user-facing CLI behavior, update the relevant docs or examples in [README.md](README.md) when needed.
- Before broad refactoring, look at the PRD and technical specification to confirm intent and avoid drifting from the product scope.

## Helpful default workflow

1. Read the task scope and relevant spec section.
2. Trace the entry point and the affected package.
3. Add or update the smallest relevant tests first.
4. Implement the root-cause fix.
5. Validate with the smallest relevant Go test or build command.
6. When an agent stops or reaches a checkpoint, run the repo’s validation commands for feedback before concluding work.

## Agent stop validation hook

The repo expects focused validation at the end of an agent run. Use the repo-standard commands as the feedback gate:

```bash
make test
make fmt
```

If a task is narrow and the repo is otherwise stable, the smallest relevant pass is acceptable, but the default expectation is a targeted validation step before the agent stops.

## Notes for AI agents

This repo is intentionally CLI-focused and safety-conscious. Favor incremental, verifiable changes over speculative refactors, and keep the user’s destructive intent explicit and reviewable.
