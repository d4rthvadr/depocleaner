# DepoCleaner 🧹

Reclaim disk space by safely identifying and reviewing large dependency directories across common ecosystems.

## Supported Ecosystems

DepoCleaner scans for dependency folders in these stacks:

- Node.js: `node_modules`
- Go: `vendor`
- PHP: `vendor`
- Python: `venv`, `.venv`
- Rust: `target`

## Features

- Smart scanning across supported ecosystems
- Fast, concurrent worker pool
- Caching to skip unchanged directories
- Interactive selection UI for safe cleanup
- Dry-run mode to preview deletions
- Clear, colorful terminal output
- Safe by design: preview, confirm, then act

## Build

Quick builds via Makefile (wraps scripts/build.sh):

```bash
# Build cross-platform binaries to dist/
make build

# Optionally set a version string
make build VERSION=v1.2.3

# Clean build artifacts
make clean
```

Artifacts are placed in `dist/` (e.g., `depo-cleaner_darwin_arm64`) with `.sha256` checksums. The build script is at [scripts/build.sh](scripts/build.sh).

## Usage

Run the CLI (defaults use config at $HOME/.depocleaner/config.yaml):

```bash
./depo-cleaner --workers 4 scan [path]
```

### Scan

```bash
# Scan your home (default from config)
./depo-cleaner scan

# Scan a specific path
./depo-cleaner scan /path/to/projects

# Control concurrency
./depo-cleaner --workers 8 scan /path/to/projects

# Disable cache for a fresh run
./depo-cleaner --workers 4 scan --no-cache /path/to/projects
```

### Clean

Interactive clean with selection UI:

```bash
# Scan and interactively select folders to delete
./depo-cleaner clean

# Clean a specific path
./depo-cleaner clean /path/to/projects

# Preview what would be deleted without actually deleting (dry run)
./depo-cleaner clean --dry-run

# Combine options
./depo-cleaner --workers 8 clean --dry-run /path/to/projects
```

The clean command workflow:

1. Scans for dependency folders
2. Presents interactive selection UI (use Space to toggle, Enter to confirm)
3. Shows confirmation prompt with total size
4. Safely deletes selected folders
5. Reports results with space reclaimed

### Config

DepoCleaner uses Viper. Configuration priority is: flags > env vars > config file > defaults.

- Config file: `$HOME/.depocleaner/config.yaml`
- Env vars prefix: `DEPOCLEANER_` (e.g., `DEPOCLEANER_WORKERS=8`)

Display current config:

```bash
./depo-cleaner config show
```

### Cache

Inspect or reset the cache:

```bash
./depo-cleaner cache clear
```

## How It Works

1. Walks directories and detects dependency folders
2. Checks cache validity by modification time
3. Analyzes size and metadata (concurrently)
4. Streams results to the UI formatter
5. Interactive selection via Bubble Tea TUI (clean command)
6. Safe deletion with confirmation and detailed reporting

## Safety First

- Preview and confirm before destructive actions
- Clear reporting of sizes and paths
- Conservative defaults for traversal depth and symlink handling

## Contributing

Pull requests are welcome. Please keep changes focused and tested.

## License

MIT

---

Made for developers who keep many projects and want clarity before cleanup.
