# NodeCleaner Technical Specification

**Version:** 1.0  
**Last Updated:** January 13, 2026  
**Status:** Draft  
**Language:** Go 1.20+  
**Target Platforms:** macOS, Linux

---

## Table of Contents

1. [Technical Overview](#1-technical-overview)
2. [System Architecture](#2-system-architecture)
3. [Technology Stack](#3-technology-stack)
4. [Data Models](#4-data-models)
5. [Core Components](#5-core-components)
6. [Algorithms & Logic](#6-algorithms--logic)
7. [File System Operations](#7-file-system-operations)
8. [Caching Strategy](#8-caching-strategy)
9. [CLI Interface Design](#9-cli-interface-design)
10. [Error Handling](#10-error-handling)
11. [Testing Strategy](#11-testing-strategy)
12. [Performance Optimization](#12-performance-optimization)
13. [Security Considerations](#13-security-considerations)
14. [Build & Distribution](#14-build--distribution)
15. [Development Workflow](#15-development-workflow)

---

## 1. Technical Overview

### 1.1 Architecture Philosophy

NodeCleaner follows a **modular, pipeline-based architecture** where data flows through discrete stages:

```
Scan → Analyze → Present → Select → Execute → Report
```

**Key Principles:**

- **Single Responsibility**: Each package handles one concern
- **Testability**: Components are mockable and unit-testable
- **Performance**: Concurrent operations where safe
- **Safety**: Multiple validation checkpoints before destructive operations

### 1.2 Project Structure

```
nodecleaner/
├── cmd/
│   └── nodecleaner/
│       └── main.go              # Application entry point
├── internal/
│   ├── scanner/
│   │   ├── scanner.go           # Filesystem scanning logic
│   │   ├── walker.go            # Directory traversal
│   │   └── filter.go            # Path filtering/ignoring
│   ├── analyzer/
│   │   ├── analyzer.go          # Metadata collection & analysis
│   │   └── stats.go             # Size calculation utilities
│   ├── cache/
│   │   ├── cache.go             # Cache management
│   │   ├── store.go             # Persistence layer
│   │   └── invalidation.go      # Cache invalidation logic
│   ├── ui/
│   │   ├── presenter.go         # Results display
│   │   ├── selector.go          # Interactive selection
│   │   └── formatter.go         # Output formatting
│   ├── cleaner/
│   │   ├── cleaner.go           # Deletion operations
│   │   └── validator.go         # Pre-deletion validation
│   ├── config/
│   │   ├── config.go            # Configuration management
│   │   └── defaults.go          # Default settings
│   └── logger/
│       └── logger.go            # Logging infrastructure
├── pkg/
│   └── models/
│       └── types.go             # Shared data types
├── test/
│   ├── integration/
│   └── fixtures/
├── scripts/
│   ├── build.sh
│   └── release.sh
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 2. System Architecture

### 2.1 Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│  (Cobra commands, flag parsing, user interaction)           │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                     Application Core                         │
│                                                               │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐    │
│  │   Scanner   │──▶│   Analyzer   │──▶│   Presenter    │    │
│  └─────────────┘  └──────────────┘  └─────────────────┘    │
│         │                                      │              │
│         │                                      │              │
│  ┌──────▼──────┐                      ┌───────▼────────┐    │
│  │    Cache    │                      │    Selector    │    │
│  └─────────────┘                      └───────┬────────┘    │
│                                               │              │
│                                       ┌───────▼────────┐    │
│                                       │    Cleaner     │    │
│                                       └────────────────┘    │
└───────────────────────────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                   Infrastructure Layer                       │
│  (Filesystem, JSON storage, logging)                        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Data Flow

**Scan Operation:**

```
User Input → Config → Scanner → Walker → Filter → Analyzer
    → Cache Check → Results → Presenter → User
```

**Clean Operation:**

```
User Selection → Validator → Cleaner → Filesystem → Logger
    → Reporter → User
```

---

## 3. Technology Stack

### 3.1 Core Dependencies

```go
// go.mod
module github.com/yourusername/nodecleaner

go 1.20

require (
    github.com/spf13/cobra v1.8.0          // CLI framework
    github.com/spf13/viper v1.18.0         // Configuration management
    github.com/charmbracelet/bubbletea v0.25.0  // TUI framework
    github.com/charmbracelet/bubbles v0.18.0    // TUI components
    github.com/charmbracelet/lipgloss v0.9.1    // Styling
    github.com/dustin/go-humanize v1.0.1   // Human-readable sizes
    github.com/schollz/progressbar/v3 v3.14.1   // Progress bars
    go.uber.org/zap v1.26.0                // Structured logging
    github.com/stretchr/testify v1.8.4     // Testing utilities
)
```

### 3.2 Standard Library Usage

- **os**: File operations, environment variables
- **path/filepath**: Cross-platform path handling
- **encoding/json**: Cache persistence
- **sync**: Concurrent operations (WaitGroup, Mutex)
- **time**: Timestamp handling
- **io/fs**: Filesystem abstraction (Go 1.16+)

### 3.3 Rationale for Key Choices

**Cobra**: Industry-standard CLI framework, excellent documentation, Git-style subcommands

**Bubbletea**: Modern TUI framework for interactive selection, clean architecture

**Zap**: High-performance structured logging with minimal allocations

**Standard Library First**: Minimize dependencies for security and binary size

---

## 4. Data Models

### 4.1 Core Types

```go
// pkg/models/types.go

package models

import "time"

// DependencyFolder represents a detected dependency directory
type DependencyFolder struct {
    Path         string    `json:"path"`
    AbsolutePath string    `json:"absolute_path"`
    Size         int64     `json:"size"`           // bytes
    ModTime      time.Time `json:"mod_time"`
    AccessTime   time.Time `json:"access_time"`
    Type         string    `json:"type"`           // "node_modules", etc.
    Selected     bool      `json:"-"`              // UI state, not persisted
}

// ScanResult encapsulates all findings from a scan
type ScanResult struct {
    Folders      []DependencyFolder `json:"folders"`
    TotalSize    int64              `json:"total_size"`
    TotalCount   int                `json:"total_count"`
    ScanPath     string             `json:"scan_path"`
    ScanTime     time.Time          `json:"scan_time"`
    Duration     time.Duration      `json:"duration"`
    CacheHits    int                `json:"cache_hits"`
    CacheMisses  int                `json:"cache_misses"`
}

// CacheEntry stores cached folder information
type CacheEntry struct {
    Path       string    `json:"path"`
    Size       int64     `json:"size"`
    ModTime    time.Time `json:"mod_time"`
    LastScan   time.Time `json:"last_scan"`
    Hash       string    `json:"hash,omitempty"`  // Optional content hash
}

// CacheIndex is the root cache structure
type CacheIndex struct {
    Version   string                `json:"version"`
    Entries   map[string]CacheEntry `json:"entries"`
    UpdatedAt time.Time             `json:"updated_at"`
}

// CleanResult records deletion outcomes
type CleanResult struct {
    Deleted       []string      `json:"deleted"`
    Failed        []FailedOp    `json:"failed"`
    SpaceReclaimed int64        `json:"space_reclaimed"`
    Duration      time.Duration `json:"duration"`
    DryRun        bool          `json:"dry_run"`
}

// FailedOp records a failed deletion
type FailedOp struct {
    Path   string `json:"path"`
    Reason string `json:"reason"`
}

// Config holds application configuration
type Config struct {
    ScanPath      string   `json:"scan_path"`
    IgnorePaths   []string `json:"ignore_paths"`
    CachePath     string   `json:"cache_path"`
    LogPath       string   `json:"log_path"`
    FollowSymlinks bool    `json:"follow_symlinks"`
    MaxDepth      int      `json:"max_depth"`
    Workers       int      `json:"workers"`
}
```

---

## 5. Core Components

### 5.1 Scanner Package

**Purpose**: Traverse filesystem and detect dependency directories

```go
// internal/scanner/scanner.go

package scanner

import (
    "context"
    "fmt"
    "path/filepath"
    "sync"
    "time"

    "github.com/yourusername/nodecleaner/pkg/models"
    "github.com/yourusername/nodecleaner/internal/analyzer"
)

type Scanner struct {
    config    *models.Config
    cache     CacheProvider
    filter    *Filter
    analyzer  *analyzer.Analyzer
    workQueue chan string
    results   chan models.DependencyFolder
    errors    chan error
}

// CacheProvider abstracts cache operations
type CacheProvider interface {
    Get(path string) (*models.CacheEntry, bool)
    Set(path string, entry models.CacheEntry) error
    IsValid(path string, modTime time.Time) bool
}

// NewScanner creates a configured scanner
func NewScanner(cfg *models.Config, cache CacheProvider) *Scanner {
    return &Scanner{
        config:    cfg,
        cache:     cache,
        filter:    NewFilter(cfg.IgnorePaths),
        analyzer:  analyzer.NewAnalyzer(),
        workQueue: make(chan string, cfg.Workers*2),
        results:   make(chan models.DependencyFolder, 100),
        errors:    make(chan error, 10),
    }
}

// Scan performs filesystem traversal
func (s *Scanner) Scan(ctx context.Context, rootPath string) (*models.ScanResult, error) {
    result := &models.ScanResult{
        ScanPath: rootPath,
        ScanTime: time.Now(),
    }

    var wg sync.WaitGroup

    // Launch worker pool
    for i := 0; i < s.config.Workers; i++ {
        wg.Add(1)
        go s.worker(ctx, &wg)
    }

    // Walk filesystem in a goroutine
    go func() {
        if err := s.walk(ctx, rootPath); err != nil {
            s.handleError(fmt.Errorf("walking filesystem: %w", err), rootPath)
        }
        close(s.workQueue) // Signal workers to finish
    }()

    // Wait for workers to complete and close channels
    go func() {
        wg.Wait()
        close(s.results)
        close(s.errors)
    }()

    // Aggregate results
    for folder := range s.results {
        result.Folders = append(result.Folders, folder)
        result.TotalSize += folder.Size
        result.TotalCount++
    }

    // Collect and handle errors using handleError method
    for err := range s.errors {
        // Log errors (could be aggregated or handled differently)
        fmt.Printf("Scan error: %v\n", err)
    }

    result.Duration = time.Since(result.ScanTime)
    return result, nil
}

// walk recursively traverses directories
func (s *Scanner) walk(ctx context.Context, rootPath string) error {
    // Implementation in section 6.1
    return nil
}

// worker processes directories from the work queue
func (s *Scanner) worker(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case path, ok := <-s.workQueue:
            if !ok {
                return // Channel closed
            }

            // Analyze the directory
            folder, err := s.analyzer.Analyze(path)
            if err != nil {
                s.errors <- fmt.Errorf("analyzing %s: %w", path, err)
                continue
            }

            // Cache the result
            if s.cache != nil {
                cacheEntry := models.CacheEntry{
                    Path:     path,
                    Size:     folder.Size,
                    ModTime:  folder.ModTime,
                    LastScan: time.Now(),
                }
                s.cache.Set(path, cacheEntry)
            }

            // Send result
            select {
            case s.results <- *folder:
            case <-ctx.Done():
                return
            }
        }
    }
}

// isTargetDir checks if directory name is a dependency folder
func (s *Scanner) isTargetDir(name string) bool {
    targetDirs := []string{
        "node_modules", // JavaScript/TypeScript
        "vendor",       // Go/PHP
        ".venv",        // Python virtual environment
        "venv",         // Python virtual environment
        "target",       // Rust/Java
    }

    for _, target := range targetDirs {
        if name == target {
            return true
        }
    }
    return false
}
```

### 5.2 Analyzer Package

**Purpose**: Collect metadata and calculate statistics

```go
// internal/analyzer/analyzer.go

package analyzer

import (
    "os"
    "path/filepath"
    "syscall"

    "github.com/yourusername/nodecleaner/pkg/models"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
    return &Analyzer{}
}

// Analyze collects metadata for a directory
func (a *Analyzer) Analyze(path string) (*models.DependencyFolder, error) {
    info, err := os.Stat(path)
    if err != nil {
        return nil, err
    }

    folder := &models.DependencyFolder{
        Path:         path,
        AbsolutePath: path,
        ModTime:      info.ModTime(),
        Type:         a.detectType(path),
    }

    // Calculate size
    size, err := a.calculateSize(path)
    if err != nil {
        return nil, err
    }
    folder.Size = size

    // Get access time (platform-specific)
    folder.AccessTime = a.getAccessTime(info)

    return folder, nil
}

// calculateSize recursively sums directory size
func (a *Analyzer) calculateSize(path string) (int64, error) {
    var size int64

    err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // Skip inaccessible files
        }
        if !info.IsDir() {
            size += info.Size()
        }
        return nil
    })

    return size, err
}

// detectType determines folder type
func (a *Analyzer) detectType(path string) string {
    base := filepath.Base(path)
    switch base {
    case "node_modules":
        return "node_modules"
    default:
        return "unknown"
    }
}

// getAccessTime extracts access time (platform-specific)
func (a *Analyzer) getAccessTime(info os.FileInfo) time.Time {
    stat := info.Sys().(*syscall.Stat_t)
    // macOS/Linux specific
    return time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
}
```

### 5.3 Cache Package

**Purpose**: Persist scan results to avoid redundant work

```go
// internal/cache/cache.go

package cache

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"

    "github.com/yourusername/nodecleaner/pkg/models"
)

type Cache struct {
    index    *models.CacheIndex
    path     string
    mu       sync.RWMutex
    modified bool
}

func NewCache(cachePath string) (*Cache, error) {
    c := &Cache{
        path: cachePath,
        index: &models.CacheIndex{
            Version: "1.0",
            Entries: make(map[string]models.CacheEntry),
        },
    }

    // Load existing cache
    if err := c.load(); err != nil && !os.IsNotExist(err) {
        return nil, err
    }

    return c, nil
}

// Get retrieves a cache entry
func (c *Cache) Get(path string) (*models.CacheEntry, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, ok := c.index.Entries[path]
    return &entry, ok
}

// Set stores a cache entry
func (c *Cache) Set(path string, entry models.CacheEntry) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.index.Entries[path] = entry
    c.modified = true
    return nil
}
```

### 7.3 Symlink Handling

```go
// internal/scanner/symlinks.go

func (s *Scanner) handleSymlink(path string) (string, error) {
    if !s.config.FollowSymlinks {
        return "", fmt.Errorf("symlinks disabled")
    }

    // Resolve symlink
    resolved, err := filepath.EvalSymlinks(path)
    if err != nil {
        return "", err
    }

    // Check for circular references
    if resolved == path {
        return "", fmt.Errorf("circular symlink detected")
    }

    return resolved, nil
}
```

---

## 8. Caching Strategy

### 8.1 Cache Structure

**File Location**: `~/.nodecleaner/cache.json`

**Format**:

```json
{
  "version": "1.0",
  "updated_at": "2026-01-13T10:30:00Z",
  "entries": {
    "/Users/dev/project1/node_modules": {
      "path": "/Users/dev/project1/node_modules",
      "size": 524288000,
      "mod_time": "2025-12-01T14:20:00Z",
      "last_scan": "2026-01-13T10:30:00Z"
    }
  }
}
```

### 8.2 Cache Operations Flow

```
┌─────────────┐
│ Scan Start  │
└──────┬──────┘
       │
       ▼
┌─────────────────┐      ┌──────────────┐
│ Load Cache File │─────▶│ Parse JSON   │
└─────────────────┘      └──────┬───────┘
                                │
                                ▼
                         ┌─────────────────┐
                         │ For Each Folder │
                         └──────┬──────────┘
                                │
                    ┌───────────┼───────────┐
                    ▼                       ▼
            ┌───────────────┐      ┌──────────────┐
            │ In Cache?     │─No──▶│ Full Scan    │
            └───────┬───────┘      └──────┬───────┘
                    │Yes                  │
                    ▼                     ▼
            ┌───────────────┐      ┌──────────────┐
            │ ModTime Same? │      │ Update Cache │
            └───────┬───────┘      └──────────────┘
                    │
        ┌───────────┼───────────┐
        │No                     │Yes
        ▼                       ▼
┌───────────────┐      ┌──────────────┐
│ Rescan Folder │      │ Use Cached   │
└───────┬───────┘      └──────────────┘
        │
        ▼
┌───────────────┐
│ Update Cache  │
└───────────────┘
```

### 8.3 Cache Invalidation Rules

```go
// internal/cache/rules.go

type InvalidationRules struct {
    MaxAge        time.Duration // 7 days default
    ForceRescan   bool
    PruneOrphans  bool
}

func (c *Cache) ApplyRules(rules InvalidationRules) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    now := time.Now()

    for path, entry := range c.index.Entries {
        shouldInvalidate := false

        // Rule 1: Age-based invalidation
        if now.Sub(entry.LastScan) > rules.MaxAge {
            shouldInvalidate = true
        }

        // Rule 2: Path no longer exists
        if rules.PruneOrphans {
            if _, err := os.Stat(path); os.IsNotExist(err) {
                delete(c.index.Entries, path)
                c.modified = true
                continue
            }
        }

        // Rule 3: Force rescan
        if rules.ForceRescan {
            shouldInvalidate = true
        }

        if shouldInvalidate {
            delete(c.index.Entries, path)
            c.modified = true
        }
    }

    return nil
}
```

---

## 9. CLI Interface Design

### 9.1 Command Structure

```
nodecleaner
├── scan [path]          # Scan for dependency folders
├── clean [path]         # Interactive clean (scan + select + delete)
├── list                 # List previous scan results
├── cache
│   ├── clear           # Clear cache
│   ├── info            # Show cache statistics
│   └── prune           # Remove orphaned entries
├── config
│   ├── show            # Display current config
│   ├── set [key=value] # Set config value
│   └── reset           # Reset to defaults
└── version             # Show version info
```

### 9.2 Command Implementation

```go
// cmd/nodecleaner/main.go

package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/yourusername/nodecleaner/internal/config"
    "github.com/yourusername/nodecleaner/internal/scanner"
    "github.com/yourusername/nodecleaner/internal/cache"
    "github.com/yourusername/nodecleaner/internal/cleaner"
    "github.com/yourusername/nodecleaner/internal/ui"
    "github.com/yourusername/nodecleaner/pkg/models"
)

var (
    cfgFile     string
    scanPath    string
    noCache     bool
    dryRun      bool
    workers     int
)

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

var rootCmd = &cobra.Command{
    Use:   "nodecleaner",
    Short: "Clean stale dependency folders",
    Long: `NodeCleaner helps developers reclaim disk space by identifying
and safely removing stale dependency directories like node_modules.`,
}

var scanCmd = &cobra.Command{
    Use:   "scan [path]",
    Short: "Scan for dependency folders",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runScan,
}

var cleanCmd = &cobra.Command{
    Use:   "clean [path]",
    Short: "Interactive clean (scan + select + delete)",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runClean,
}

var cacheCmd = &cobra.Command{
    Use:   "cache",
    Short: "Manage cache",
}

var cacheClearCmd = &cobra.Command{
    Use:   "clear",
    Short: "Clear cache",
    RunE:  runCacheClear,
}

var cacheInfoCmd = &cobra.Command{
    Use:   "info",
    Short: "Show cache statistics",
    RunE:  runCacheInfo,
}

func init() {
    cobra.OnInitialize(initConfig)

    // Global flags
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.nodecleaner/config.yaml)")
    rootCmd.PersistentFlags().IntVar(&workers, "workers", 4, "number of worker goroutines")

    // Scan flags
    scanCmd.Flags().BoolVar(&noCache, "no-cache", false, "disable cache")
    scanCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: $HOME)")

    // Clean flags
    cleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview without deleting")
    cleanCmd.Flags().BoolVar(&noCache, "no-cache", false, "disable cache")
    cleanCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: $HOME)")

    // Add commands
    rootCmd.AddCommand(scanCmd)
    rootCmd.AddCommand(cleanCmd)
    rootCmd.AddCommand(cacheCmd)

    cacheCmd.AddCommand(cacheClearCmd)
    cacheCmd.AddCommand(cacheInfoCmd)
}

func initConfig() {
    // Initialize configuration
    config.Init(cfgFile)
}

func runScan(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    // Determine scan path
    path := scanPath
    if len(args) > 0 {
        path = args[0]
    }
    if path == "" {
        path = os.Getenv("HOME")
    }

    // Load config
    cfg := config.Load()
    cfg.ScanPath = path
    cfg.Workers = workers

    // Initialize cache
    var c *cache.Cache
    var err error
    if !noCache {
        c, err = cache.NewCache(cfg.CachePath)
        if err != nil {
            return fmt.Errorf("initializing cache: %w", err)
        }
        defer c.Save()
    }

    // Create scanner
    s := scanner.NewScanner(cfg, c)

    // Run scan
    fmt.Printf("Scanning %s...\n", path)
    result, err := s.Scan(ctx, path)
    if err != nil {
        return fmt.Errorf("scanning: %w", err)
    }

    // Display results
    ui.DisplayScanResults(result)

    return nil
}

func runClean(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    // First, perform scan
    path := scanPath
    if len(args) > 0 {
        path = args[0]
    }
    if path == "" {
        path = os.Getenv("HOME")
    }

    cfg := config.Load()
    cfg.ScanPath = path
    cfg.Workers = workers

    var c *cache.Cache
    var err error
    if !noCache {
        c, err = cache.NewCache(cfg.CachePath)
        if err != nil {
            return fmt.Errorf("initializing cache: %w", err)
        }
        defer c.Save()
    }

    s := scanner.NewScanner(cfg, c)

    fmt.Printf("Scanning %s...\n", path)
    result, err := s.Scan(ctx, path)
    if err != nil {
        return fmt.Errorf("scanning: %w", err)
    }

    if len(result.Folders) == 0 {
        fmt.Println("No dependency folders found.")
        return nil
    }

    // Interactive selection
    model := ui.NewSelectionModel(result.Folders)
    p := tea.NewProgram(model)

    finalModel, err := p.Run()
    if err != nil {
        return fmt.Errorf("selection UI: %w", err)
    }

    selected := finalModel.(ui.SelectionModel).GetSelected()

    if len(selected) == 0 {
        fmt.Println("No folders selected.")
        return nil
    }

    // Confirm deletion
    if !dryRun {
        fmt.Printf("\nAre you sure you want to delete %d folders? (yes/no): ", len(selected))
        var confirm string
        fmt.Scanln(&confirm)

        if confirm != "yes" {
            fmt.Println("Deletion cancelled.")
            return nil
        }
    }

    // Perform deletion
    logger := initLogger(cfg.LogPath)
    cl := cleaner.NewCleaner(dryRun, logger)

    cleanResult, err := cl.Clean(ctx, selected)
    if err != nil {
        return fmt.Errorf("cleaning: %w", err)
    }

    // Display results
    ui.DisplayCleanResults(cleanResult)

    return nil
}

// initLogger initializes the zap structured logger
func initLogger(logPath string) *zap.SugaredLogger {
    // Ensure log directory exists
    logDir := filepath.Dir(logPath)
    if err := os.MkdirAll(logDir, 0755); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to create log directory: %v\n", err)
        // Fall back to console-only logging
        logger, _ := zap.NewProduction()
        return logger.Sugar()
    }

    // Configure zap logger
    cfg := zap.NewProductionConfig()
    cfg.OutputPaths = []string{
        logPath,      // Write to file
        "stdout",     // Also write to console
    }
    cfg.ErrorOutputPaths = []string{
        logPath,
        "stderr",
    }
    cfg.Encoding = "json"
    cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

    logger, err := cfg.Build()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to initialize logger: %v\n", err)
        logger, _ = zap.NewProduction()
    }

    return logger.Sugar()
}

func runCacheClear(cmd *cobra.Command, args []string) error {
    cfg := config.Load()
    c, err := cache.NewCache(cfg.CachePath)
    if err != nil {
        return err
    }

    if err := c.Clear(); err != nil {
        return err
    }

    fmt.Println("Cache cleared successfully.")
    return nil
}

func runCacheInfo(cmd *cobra.Command, args []string) error {
    cfg := config.Load()
    c, err := cache.NewCache(cfg.CachePath)
    if err != nil {
        return err
    }

    info := c.GetInfo()
    fmt.Printf("Cache entries: %d\n", info.EntryCount)
    fmt.Printf("Total cached size: %s\n", humanize.Bytes(uint64(info.TotalSize)))
    fmt.Printf("Last updated: %s\n", info.UpdatedAt.Format(time.RFC3339))

    return nil
}
```

### 9.3 Output Formatting

```go
// internal/ui/formatter.go

package ui

import (
    "fmt"
    "os"
    "text/tabwriter"

    "github.com/dustin/go-humanize"
    "github.com/charmbracelet/lipgloss"

    "github.com/yourusername/nodecleaner/pkg/models"
)

var (
    headerStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("12"))

    successStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("10"))

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("9"))

    warningStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("11"))
)

func DisplayScanResults(result *models.ScanResult) {
    fmt.Println(headerStyle.Render("\n📊 Scan Results"))
    fmt.Println(strings.Repeat("─", 80))

    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    fmt.Fprintln(w, "SIZE\tLAST ACCESSED\tPATH")
    fmt.Fprintln(w, strings.Repeat("─", 80))

    for _, folder := range result.Folders {
        fmt.Fprintf(w, "%s\t%s\t%s\n",
            humanize.Bytes(uint64(folder.Size)),
            humanize.Time(folder.AccessTime),
            folder.Path,
        )
    }

    w.Flush()

    fmt.Println(strings.Repeat("─", 80))
    fmt.Printf("\n%s\n", headerStyle.Render("Summary:"))
    fmt.Printf("  Total folders: %d\n", result.TotalCount)
    fmt.Printf("  Total size: %s\n", humanize.Bytes(uint64(result.TotalSize)))
    fmt.Printf("  Scan duration: %s\n", result.Duration)

    if result.CacheHits > 0 {
        fmt.Printf("  Cache hits: %d (%.1f%%)\n",
            result.CacheHits,
            float64(result.CacheHits)/float64(result.CacheHits+result.CacheMisses)*100)
    }
}

func DisplayCleanResults(result *models.CleanResult) {
    fmt.Println()

    if result.DryRun {
        fmt.Println(warningStyle.Render("🔍 DRY RUN MODE - No files were deleted"))
        fmt.Println()
    }

    fmt.Println(headerStyle.Render("🧹 Clean Results"))
    fmt.Println(strings.Repeat("─", 80))

    if len(result.Deleted) > 0 {
        fmt.Println(successStyle.Render(fmt.Sprintf("\n✓ Successfully deleted %d folders:", len(result.Deleted))))
        for _, path := range result.Deleted {
            fmt.Printf("  • %s\n", path)
        }
    }

    if len(result.Failed) > 0 {
        fmt.Println(errorStyle.Render(fmt.Sprintf("\n✗ Failed to delete %d folders:", len(result.Failed))))
        for _, fail := range result.Failed {
            fmt.Printf("  • %s: %s\n", fail.Path, fail.Reason)
        }
    }

    fmt.Println(strings.Repeat("─", 80))
    fmt.Printf("\n%s\n", headerStyle.Render("Summary:"))
    fmt.Printf("  Space reclaimed: %s\n", successStyle.Render(humanize.Bytes(uint64(result.SpaceReclaimed))))
    fmt.Printf("  Duration: %s\n", result.Duration)

    if !result.DryRun {
        fmt.Printf("\n%s\n", successStyle.Render("✓ Cleanup complete!"))
    }
}
```

---

## 10. Error Handling

### 10.1 Error Types

```go
// pkg/models/errors.go

package models

import "fmt"

// Error types
type ErrorType string

const (
    ErrTypePermission   ErrorType = "PERMISSION_DENIED"
    ErrTypeNotFound     ErrorType = "NOT_FOUND"
    ErrTypeInvalidPath  ErrorType = "INVALID_PATH"
    ErrTypeCacheCorrupt ErrorType = "CACHE_CORRUPT"
    ErrTypeIO           ErrorType = "IO_ERROR"
)

// ApplicationError represents a structured error
type ApplicationError struct {
    Type    ErrorType
    Message string
    Path    string
    Err     error
}

func (e *ApplicationError) Error() string {
    if e.Path != "" {
        return fmt.Sprintf("%s: %s (path: %s)", e.Type, e.Message, e.Path)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *ApplicationError) Unwrap() error {
    return e.Err
}

// Error constructors
func NewPermissionError(path string, err error) *ApplicationError {
    return &ApplicationError{
        Type:    ErrTypePermission,
        Message: "permission denied",
        Path:    path,
        Err:     err,
    }
}

func NewNotFoundError(path string) *ApplicationError {
    return &ApplicationError{
        Type:    ErrTypeNotFound,
        Message: "path not found",
        Path:    path,
    }
}
```

### 10.2 Error Handling Strategy

```go
// internal/scanner/errors.go

func (s *Scanner) handleError(err error, path string) {
    var appErr *models.ApplicationError

    if errors.As(err, &appErr) {
        // Structured error - log appropriately
        switch appErr.Type {
        case models.ErrTypePermission:
            s.logger.Warn("Skipping inaccessible path", "path", path)
        case models.ErrTypeNotFound:
            s.logger.Debug("Path not found", "path", path)
        default:
            s.logger.Error("Scan error", "error", appErr, "path", path)
        }
    } else {
        // Unexpected error - log with full context
        s.logger.Error("Unexpected error", "error", err, "path", path)
    }

    // Send to error channel for aggregation
    select {
    case s.errors <- err:
    default:
        // Channel full, drop error (already logged)
    }
}
```

### 10.3 Recovery Mechanisms

```go
// internal/cleaner/recovery.go

// RecoveryLog stores deletion history for potential recovery
type RecoveryLog struct {
    path string
    mu   sync.Mutex
}

func NewRecoveryLog(path string) *RecoveryLog {
    return &RecoveryLog{path: path}
}

func (r *RecoveryLog) Record(op DeletionOp) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    f, err := os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    entry := fmt.Sprintf("%s|%s|%d|%s\n",
        time.Now().Format(time.RFC3339),
        op.Path,
        op.Size,
        op.Status,
    )

    _, err = f.WriteString(entry)
    return err
}
```

---

## 11. Testing Strategy

### 11.1 Unit Tests

```go
// internal/scanner/scanner_test.go

package scanner

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/yourusername/nodecleaner/pkg/models"
)

func TestScanner_Scan(t *testing.T) {
    // Setup test directory structure
    tmpDir := t.TempDir()

    // Create test folders
    testPaths := []string{
        filepath.Join(tmpDir, "project1", "node_modules"),
        filepath.Join(tmpDir, "project2", "node_modules"),
        filepath.Join(tmpDir, "project3", "src", "node_modules"),
    }

    for _, p := range testPaths {
        require.NoError(t, os.MkdirAll(p, 0755))

        // Add some files to create size
        testFile := filepath.Join(p, "test.txt")
        require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))
    }

    // Create scanner
    cfg := &models.Config{
        ScanPath: tmpDir,
        Workers:  2,
    }
    s := NewScanner(cfg, nil)

    // Run scan
    ctx := context.Background()
    result, err := s.Scan(ctx, tmpDir)

    // Assertions
    require.NoError(t, err)
    assert.Equal(t, 3, result.TotalCount)
    assert.Greater(t, result.TotalSize, int64(0))

    // Verify all paths found
    foundPaths := make(map[string]bool)
    for _, folder := range result.Folders {
        foundPaths[folder.Path] = true
    }

    for _, expected := range testPaths {
        assert.True(t, foundPaths[expected], "Expected path not found: %s", expected)
    }
}

func TestScanner_IgnoresPaths(t *testing.T) {
    tmpDir := t.TempDir()

    // Create both valid and ignored paths
    validPath := filepath.Join(tmpDir, "project", "node_modules")
    ignoredPath := filepath.Join(tmpDir, ".hidden", "node_modules")

    require.NoError(t, os.MkdirAll(validPath, 0755))
    require.NoError(t, os.MkdirAll(ignoredPath, 0755))

    cfg := &models.Config{
        ScanPath: tmpDir,
        Workers:  1,
    }
    s := NewScanner(cfg, nil)

    ctx := context.Background()
    result, err := s.Scan(ctx, tmpDir)

    require.NoError(t, err)
    assert.Equal(t, 1, result.TotalCount)
    assert.Equal(t, validPath, result.Folders[0].Path)
}
```

### 11.2 Integration Tests

```go
// test/integration/scan_clean_test.go

package integration

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/yourusername/nodecleaner/internal/scanner"
    "github.com/yourusername/nodecleaner/internal/cleaner"
    "github.com/yourusername/nodecleaner/internal/cache"
    "github.com/yourusername/nodecleaner/pkg/models"
)

func TestFullScanAndCleanWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    tmpDir := t.TempDir()
    cacheDir := t.TempDir()

    // Setup test structure
    testFolder := filepath.Join(tmpDir, "test-project", "node_modules")
    require.NoError(t, os.MkdirAll(testFolder, 0755))

    // Add content
    for i := 0; i < 10; i++ {
        file := filepath.Join(testFolder, fmt.Sprintf("file%d.txt", i))
        content := make([]byte, 1024*100) // 100KB each
        require.NoError(t, os.WriteFile(file, content, 0644))
    }

    // Initialize components
    cfg := &models.Config{
        ScanPath:  tmpDir,
        CachePath: filepath.Join(cacheDir, "cache.json"),
        Workers:   2,
    }

    c, err := cache.NewCache(cfg.CachePath)
    require.NoError(t, err)

    s := scanner.NewScanner(cfg, c)

    // Scan
    ctx := context.Background()
    scanResult, err := s.Scan(ctx, tmpDir)
    require.NoError(t, err)
    assert.Equal(t, 1, scanResult.TotalCount)

    // Save cache
    require.NoError(t, c.Save())

    // Clean
    logger := &testLogger{}
    cl := cleaner.NewCleaner(false, logger)

    cleanResult, err := cl.Clean(ctx, scanResult.Folders)
    require.NoError(t, err)

    assert.Equal(t, 1, len(cleanResult.Deleted))
    assert.Equal(t, 0, len(cleanResult.Failed))
    assert.Greater(t, cleanResult.SpaceReclaimed, int64(0))

    // Verify deletion
    _, err = os.Stat(testFolder)
    assert.True(t, os.IsNotExist(err), "Folder should be deleted")
}

type testLogger struct{}

func (l *testLogger) Info(msg string, fields ...interface{})  {}
func (l *testLogger) Error(msg string, fields ...interface{}) {}
func (l *testLogger) Warn(msg string, fields ...interface{})  {}
```

### 11.3 Benchmark Tests

```go
// internal/scanner/scanner_bench_test.go

package scanner

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/yourusername/nodecleaner/pkg/models"
)

func BenchmarkScanner_Scan(b *testing.B) {
    tmpDir := b.TempDir()

    // Create 100 node_modules folders
    for i := 0; i < 100; i++ {
        path := filepath.Join(tmpDir, fmt.Sprintf("project%d", i), "node_modules")
        os.MkdirAll(path, 0755)

        // Add some files
        for j := 0; j < 10; j++ {
            file := filepath.Join(path, fmt.Sprintf("file%d.txt", j))
            os.WriteFile(file, []byte("test"), 0644)
        }
    }

    cfg := &models.Config{
        ScanPath: tmpDir,
        Workers:  4,
    }

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        s := NewScanner(cfg, nil)
        ctx := context.Background()
        _, err := s.Scan(ctx, tmpDir)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## 12. Performance Optimization

### 12.1 Concurrency Strategy

**Worker Pool Pattern**:

```go
// internal/scanner/pool.go

type WorkerPool struct {
    workers int
    jobs    chan string
    results chan models.DependencyFolder
    wg      sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        jobs:    make(chan string, workers*2),
        results: make(chan models.DependencyFolder, 100),
    }
}

func (p *WorkerPool) Start(ctx context.Context, analyzer *analyzer.Analyzer) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, analyzer)
    }
}

func (p *WorkerPool) worker(ctx context.Context, a *analyzer.Analyzer) {
    defer p.wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case path, ok := <-p.jobs:
            if !ok {
                return
            }

            folder, err := a.Analyze(path)
            if err != nil {
                continue
            }

            select {
            case p.results <- *folder:
            case <-ctx.Done():
                return
            }
        }
    }
}

func (p *WorkerPool) Submit(path string) {
    p.jobs <- path
}

func (p *WorkerPool) Close() {
    close(p.jobs)
    p.wg.Wait()
    close(p.results)
}
```

### 12.2 Memory Management

```go
// internal/scanner/memory.go

// Streaming results to avoid loading everything in memory
type StreamingScanner struct {
    resultWriter ResultWriter
}

type ResultWriter interface {
    Write(folder models.DependencyFolder) error
    Flush() error
}

// File-based result writer for large scans
type FileResultWriter struct {
    file *os.File
    enc  *json.Encoder
}

func NewFileResultWriter(path string) (*FileResultWriter, error) {
    f, err := os.Create(path)
    if err != nil {
        return nil, err
    }

    return &FileResultWriter{
        file: f,
        enc:  json.NewEncoder(f),
    }, nil
}

func (w *FileResultWriter) Write(folder models.DependencyFolder) error {
    return w.enc.Encode(folder)
}

func (w *FileResultWriter) Flush() error {
    return w.file.Sync()
}

func (w *FileResultWriter) Close() error {
    return w.file.Close()
}
```

### 12.3 Optimization Techniques

**1. Directory Skip Optimization**:

```go

// IsValid checks if cached entry is still valid
func (c *Cache) IsValid(path string, currentModTime time.Time) bool {
    entry, ok := c.Get(path)
    if !ok {
        return false
    }

    // Invalid if modification time changed
    return entry.ModTime.Equal(currentModTime)
}

// Save persists cache to disk
func (c *Cache) Save() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if !c.modified {
        return nil
    }

    c.index.UpdatedAt = time.Now()

    // Ensure directory exists
    dir := filepath.Dir(c.path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Write to temp file, then rename (atomic)
    tempPath := c.path + ".tmp"
    f, err := os.Create(tempPath)
    if err != nil {
        return err
    }

    encoder := json.NewEncoder(f)
    encoder.SetIndent("", "  ")
    err = encoder.Encode(c.index)
    f.Close()

    if err != nil {
        os.Remove(tempPath)
        return err
    }

    return os.Rename(tempPath, c.path)
}

// load reads cache from disk
func (c *Cache) load() error {
    f, err := os.Open(c.path)
    if err != nil {
        return err
    }
    defer f.Close()

    return json.NewDecoder(f).Decode(&c.index)
}

// Clear removes all cache entries
func (c *Cache) Clear() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.index.Entries = make(map[string]models.CacheEntry)
    c.modified = true
    return c.Save()
}

// Prune removes cache entries for non-existent paths
func (c *Cache) Prune() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    for path := range c.index.Entries {
        if _, err := os.Stat(path); os.IsNotExist(err) {
            delete(c.index.Entries, path)
            c.modified = true
        }
    }

    return nil
}
```

### 5.4 UI Package

**Purpose**: Interactive terminal interface for selection

```go
// internal/ui/selector.go

package ui

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/lipgloss"
    "github.com/dustin/go-humanize"

    "github.com/yourusername/nodecleaner/pkg/models"
)

type SelectionModel struct {
    table         table.Model
    folders       []models.DependencyFolder
    selected      map[int]bool
    totalSelected int64
}

func NewSelectionModel(folders []models.DependencyFolder) SelectionModel {
    // Create table columns
    columns := []table.Column{
        {Title: "✓", Width: 3},
        {Title: "Size", Width: 10},
        {Title: "Last Accessed", Width: 12},
        {Title: "Path", Width: 60},
    }

    // Convert folders to rows
    rows := make([]table.Row, len(folders))
    for i, f := range folders {
        rows[i] = table.Row{
            "",
            humanize.Bytes(uint64(f.Size)),
            humanize.Time(f.AccessTime),
            f.Path,
        }
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(20),
    )

    return SelectionModel{
        table:    t,
        folders:  folders,
        selected: make(map[int]bool),
    }
}

func (m SelectionModel) Init() tea.Cmd {
    return nil
}

func (m SelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case " ":
            // Toggle selection
            idx := m.table.Cursor()
            m.selected[idx] = !m.selected[idx]

            // Update total
            if m.selected[idx] {
                m.totalSelected += m.folders[idx].Size
            } else {
                m.totalSelected -= m.folders[idx].Size
            }

            // Update row
            m.updateRow(idx)

        case "enter":
            // Confirm selection
            return m, tea.Quit
        }
    }

    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m SelectionModel) View() string {
    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("12")).
        Render(fmt.Sprintf("Select folders to delete (Space to toggle, Enter to confirm)\nSelected: %s\n\n",
            humanize.Bytes(uint64(m.totalSelected))))

    return header + m.table.View() + "\n"
}

func (m *SelectionModel) updateRow(idx int) {
    checkmark := ""
    if m.selected[idx] {
        checkmark = "✓"
    }

    f := m.folders[idx]
    m.table.SetRows([]table.Row{
        {
            checkmark,
            humanize.Bytes(uint64(f.Size)),
            humanize.Time(f.AccessTime),
            f.Path,
        },
    })
}

// GetSelected returns selected folders
func (m SelectionModel) GetSelected() []models.DependencyFolder {
    var selected []models.DependencyFolder
    for i, isSelected := range m.selected {
        if isSelected {
            selected = append(selected, m.folders[i])
        }
    }
    return selected
}
```

### 5.5 Cleaner Package

**Purpose**: Safe deletion operations

```go
// internal/cleaner/cleaner.go

package cleaner

import (
    "context"
    "os"
    "sync"

    "github.com/yourusername/nodecleaner/pkg/models"
)

type Cleaner struct {
    dryRun bool
    logger Logger
}

type Logger interface {
    Info(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
}

func NewCleaner(dryRun bool, logger Logger) *Cleaner {
    return &Cleaner{
        dryRun: dryRun,
        logger: logger,
    }
}

// Clean removes selected folders
func (c *Cleaner) Clean(ctx context.Context, folders []models.DependencyFolder) (*models.CleanResult, error) {
    result := &models.CleanResult{
        DryRun: c.dryRun,
    }

    // Note: mu and wg are local to this function call, not struct fields.
    // This allows multiple Clean operations to run concurrently without
    // interfering with each other. Each operation gets its own synchronization primitives.
    var mu sync.Mutex
    var wg sync.WaitGroup

    // Process deletions concurrently
    for _, folder := range folders {
        wg.Add(1)
        go func(f models.DependencyFolder) {
            defer wg.Done()

            if err := c.deleteFolder(ctx, f.Path); err != nil {
                mu.Lock()
                result.Failed = append(result.Failed, models.FailedOp{
                    Path:   f.Path,
                    Reason: err.Error(),
                })
                mu.Unlock()
                c.logger.Error("Failed to delete", "path", f.Path, "error", err)
            } else {
                mu.Lock()
                result.Deleted = append(result.Deleted, f.Path)
                result.SpaceReclaimed += f.Size
                mu.Unlock()
                c.logger.Info("Deleted", "path", f.Path, "size", f.Size)
            }
        }(folder)
    }

    wg.Wait()
    return result, nil
}

// deleteFolder removes a directory
func (c *Cleaner) deleteFolder(ctx context.Context, path string) error {
    // Validate path still exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("path no longer exists")
    }

    if c.dryRun {
        c.logger.Info("DRY RUN: Would delete", "path", path)
        return nil
    }

    // Check context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    return os.RemoveAll(path)
}
```

### 5.6 Config Package

**Purpose**: Configuration management using Viper

```go
// internal/config/config.go

package config

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/viper"
    "github.com/yourusername/nodecleaner/pkg/models"
)

var (
    globalConfig *models.Config
    configPath   string
)

// Init initializes Viper configuration
func Init(cfgFile string) error {
    if cfgFile != "" {
        // Use config file from flag
        viper.SetConfigFile(cfgFile)
    } else {
        // Search for config in standard locations
        home, err := os.UserHomeDir()
        if err != nil {
            return fmt.Errorf("getting home directory: %w", err)
        }

        configDir := filepath.Join(home, ".nodecleaner")

        // Create config directory if it doesn't exist
        if err := os.MkdirAll(configDir, 0755); err != nil {
            return fmt.Errorf("creating config directory: %w", err)
        }

        viper.AddConfigPath(configDir)
        viper.AddConfigPath(".")
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")

        configPath = filepath.Join(configDir, "config.yaml")
    }

    // Set defaults
    setDefaults()

    // Enable environment variable support
    viper.SetEnvPrefix("NODECLEANER")
    viper.AutomaticEnv()

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found; create with defaults
            if err := SaveDefaults(); err != nil {
                return fmt.Errorf("creating default config: %w", err)
            }
        } else {
            return fmt.Errorf("reading config: %w", err)
        }
    }

    // Unmarshal into global config
    globalConfig = &models.Config{}
    if err := viper.Unmarshal(globalConfig); err != nil {
        return fmt.Errorf("unmarshaling config: %w", err)
    }

    return nil
}

// setDefaults sets default configuration values
func setDefaults() {
    home, _ := os.UserHomeDir()
    configDir := filepath.Join(home, ".nodecleaner")

    viper.SetDefault("scan_path", home)
    viper.SetDefault("cache_path", filepath.Join(configDir, "cache.json"))
    viper.SetDefault("log_path", filepath.Join(configDir, "nodecleaner.log"))
    viper.SetDefault("follow_symlinks", false)
    viper.SetDefault("max_depth", 10)
    viper.SetDefault("workers", 4)
    viper.SetDefault("ignore_paths", []string{
        "/System",
        "/Library",
        "/Applications",
        "/private/var",
        "/dev",
        "/proc",
        "/sys",
        "/.Trash",
        "/Network",
    })
}

// Load returns the current configuration
func Load() *models.Config {
    if globalConfig == nil {
        globalConfig = &models.Config{
            Workers:  viper.GetInt("workers"),
            MaxDepth: viper.GetInt("max_depth"),
        }
        home, _ := os.UserHomeDir()
        configDir := filepath.Join(home, ".nodecleaner")
        globalConfig.CachePath = filepath.Join(configDir, "cache.json")
        globalConfig.LogPath = filepath.Join(configDir, "nodecleaner.log")
    }
    return globalConfig
}

// Set updates a configuration value
func Set(key string, value interface{}) error {
    viper.Set(key, value)

    // Update global config
    if globalConfig != nil {
        if err := viper.Unmarshal(globalConfig); err != nil {
            return fmt.Errorf("updating config: %w", err)
        }
    }

    return nil
}

// Save persists current configuration to disk
func Save() error {
    if configPath == "" {
        home, _ := os.UserHomeDir()
        configPath = filepath.Join(home, ".nodecleaner", "config.yaml")
    }

    return viper.WriteConfigAs(configPath)
}

// SaveDefaults creates a config file with default values
func SaveDefaults() error {
    home, _ := os.UserHomeDir()
    configDir := filepath.Join(home, ".nodecleaner")
    configPath = filepath.Join(configDir, "config.yaml")

    if err := os.MkdirAll(configDir, 0755); err != nil {
        return err
    }

    return viper.SafeWriteConfigAs(configPath)
}

// GetString retrieves a string configuration value
func GetString(key string) string {
    return viper.GetString(key)
}

// GetInt retrieves an integer configuration value
func GetInt(key string) int {
    return viper.GetInt(key)
}

// GetBool retrieves a boolean configuration value
func GetBool(key string) bool {
    return viper.GetBool(key)
}

// GetStringSlice retrieves a string slice configuration value
func GetStringSlice(key string) []string {
    return viper.GetStringSlice(key)
}

// Display prints current configuration
func Display() {
    fmt.Println("Current Configuration:")
    fmt.Println("=====================")

    allSettings := viper.AllSettings()
    for key, value := range allSettings {
        fmt.Printf("%s: %v\n", key, value)
    }
}

// Reset restores default configuration
func Reset() error {
    // Clear all settings
    for key := range viper.AllSettings() {
        viper.Set(key, nil)
    }

    // Reapply defaults
    setDefaults()

    // Save to file
    return SaveDefaults()
}
```

### 5.7 Logger Package

**Purpose**: Structured logging with Zap

```go
// internal/logger/logger.go

package logger

import (
    "fmt"
    "os"
    "path/filepath"

    "go.uber.org/zap"g
    "go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger
type Logger struct {
    *zap.SugaredLogger
}

// New creates a new logger instance
func New(logPath string, level string) (*Logger, error) {
    // Ensure log directory exists
    logDir := filepath.Dir(logPath)
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return nil, fmt.Errorf("creating log directory: %w", err)
    }

    // Parse log level
    logLevel, err := zapcore.ParseLevel(level)
    if err != nil {
        logLevel = zapcore.InfoLevel
    }

    // Configure zap logger
    cfg := zap.NewProductionConfig()
    cfg.OutputPaths = []string{
        logPath,  // Write to file
        "stdout", // Also write to console
    }
    cfg.ErrorOutputPaths = []string{
        logPath,
        "stderr",
    }
    cfg.Encoding = "json"
    cfg.Level = zap.NewAtomicLevelAt(logLevel)

    // Customize time encoding for better readability
    cfg.EncoderConfig.TimeKey = "timestamp"
    cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    logger, err := cfg.Build()
    if err != nil {
        return nil, fmt.Errorf("building logger: %w", err)
    }

    return &Logger{logger.Sugar()}, nil
}

// NewConsoleOnly creates a console-only logger (for testing)
func NewConsoleOnly() *Logger {
    logger, _ := zap.NewDevelopment()
    return &Logger{logger.Sugar()}
}

// WithFields returns a logger with additional context fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
    var zapFields []interface{}
    for k, v := range fields {
        zapFields = append(zapFields, k, v)
    }
    return &Logger{l.SugaredLogger.With(zapFields...)}
}

// Close flushes any buffered log entries
func (l *Logger) Close() error {
    return l.Sync()
}
```

**Logger Interface Adapter**:

```go
// internal/logger/adapter.go

package logger

// Adapter makes Logger compatible with cleaner.Logger interface
type Adapter struct {
    logger *Logger
}

// NewAdapter creates a logger adapter
func NewAdapter(logger *Logger) *Adapter {
    return &Adapter{logger: logger}
}

// Info logs an info message
func (a *Adapter) Info(msg string, fields ...interface{}) {
    a.logger.Infow(msg, fields...)
}

// Error logs an error message
func (a *Adapter) Error(msg string, fields ...interface{}) {
    a.logger.Errorw(msg, fields...)
}

// Warn logs a warning message
func (a *Adapter) Warn(msg string, fields ...interface{}) {
    a.logger.Warnw(msg, fields...)
}
```

### 5.8 Logger Usage Guidelines

**When to Use the Adapter:**

The Logger Adapter is used to bridge different logging interface requirements across components:

**Scanner Component:**

- Uses `logger.Logger` directly (returns `*Logger` from `New()`)
- Scanner.Logger interface requires: `Info`, `Error`, `Warn`, `Debug` methods
- Pass the logger directly without wrapping

```go
// In command implementation
logger, err := logger.New(cfg.LogPath, "info")
if err != nil {
    return err
}
defer logger.Close()

// Pass directly to Scanner
scanner := scanner.NewScanner(cfg, cache, logger)
```

**Cleaner Component:**

- Uses `logger.Adapter` to wrap the logger
- Cleaner.Logger interface requires only: `Info`, `Error` methods
- Wrap with `NewAdapter()` for interface compatibility

```go
// Create base logger
logger, err := logger.New(cfg.LogPath, "info")
if err != nil {
    return err
}

// Wrap for Cleaner
cleanerLogger := logger.NewAdapter(logger)
cleaner := cleaner.NewCleaner(dryRun, cleanerLogger)
```

**Testing:**

- Create mock loggers that implement the required interface
- Use Adapter when needed for interface compatibility

```go
type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields ...interface{})  {}
func (m *mockLogger) Error(msg string, fields ...interface{}) {}
func (m *mockLogger) Warn(msg string, fields ...interface{})  {}

// Use directly or wrap with Adapter as needed
```

**Key Design Benefits:**

1. **Interface Segregation**: Components only depend on methods they actually use
2. **Flexibility**: Easy to swap implementations without affecting components
3. **Testability**: Mock loggers can be minimal and focused
4. **Maintainability**: Changes to logging don't cascade through all components

**Configuration File Example** (`~/.nodecleaner/config.yaml`):

```yaml
# NodeCleaner Configuration

# Default scan path (defaults to $HOME if not specified)
scan_path: /Users/username

# Cache file location
cache_path: /Users/username/.nodecleaner/cache.json

# Log file location
log_path: /Users/username/.nodecleaner/nodecleaner.log

# Follow symbolic links during scan
follow_symlinks: false

# Maximum directory depth to traverse (0 = unlimited)
max_depth: 10

# Number of concurrent worker goroutines
workers: 4

# Paths to ignore during scanning
ignore_paths:
  - /System
  - /Library
  - /Applications
  - /private/var
  - /dev
  - /proc
  - /sys
  - /.Trash
  - /Network
```

### 5.9 Configuration Priority

Viper follows this configuration priority (highest to lowest):

1. **Command-line flags**: `--workers 8`
2. **Environment variables**: `NODECLEANER_WORKERS=8`
3. **Config file**: `config.yaml`
4. **Default values**: Built-in defaults

**Example Environment Variables**:

```bash
# Set custom scan path
export NODECLEANER_SCAN_PATH=/Users/dev/projects

# Set number of workers
export NODECLEANER_WORKERS=8

# Disable symlink following
export NODECLEANER_FOLLOW_SYMLINKS=false

# Run with environment variables
nodecleaner scan
```

### 5.10 Config Management Commands

```go
// cmd/nodecleaner/config_commands.go

var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
    Use:   "show",
    Short: "Display current configuration",
    RunE: func(cmd *cobra.Command, args []string) error {
        config.Display()
        return nil
    },
}

var configSetCmd = &cobra.Command{
    Use:   "set [key=value]",
    Short: "Set configuration value",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        parts := strings.SplitN(args[0], "=", 2)
        if len(parts) != 2 {
            return fmt.Errorf("invalid format, use key=value")
        }

        key, value := parts[0], parts[1]

        // Parse value based on expected type
        if err := config.Set(key, parseValue(value)); err != nil {
            return err
        }

        if err := config.Save(); err != nil {
            return fmt.Errorf("saving config: %w", err)
        }

        fmt.Printf("Set %s = %s\n", key, value)
        return nil
    },
}

var configResetCmd = &cobra.Command{
    Use:   "reset",
    Short: "Reset configuration to defaults",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := config.Reset(); err != nil {
            return fmt.Errorf("resetting config: %w", err)
        }

        fmt.Println("Configuration reset to defaults")
        return nil
    },
}

func init() {
    configCmd.AddCommand(configShowCmd)
    configCmd.AddCommand(configSetCmd)
    configCmd.AddCommand(configResetCmd)
    rootCmd.AddCommand(configCmd)
}

// parseValue attempts to parse string value to appropriate type
func parseValue(value string) interface{} {
    // Try boolean
    if value == "true" || value == "false" {
        return value == "true"
    }

    // Try integer
    if i, err := strconv.Atoi(value); err == nil {
        return i
    }

    // Default to string
    return value
}
```

---

## 6. Algorithms & Logic

### 6.1 Filesystem Traversal Algorithm

**Approach**: Depth-First Search using `filepath.WalkDir` (Go 1.16+)

**Why `filepath.WalkDir`:**

- More efficient than manual recursion (uses `fs.DirEntry` avoiding extra syscalls)
- Built-in depth-first traversal
- Better error handling with `SkipDir` and `SkipAll`
- Standard library implementation (battle-tested)
- Lower memory footprint

```go
// internal/scanner/walker.go

package scanner

import (
    "context"
    "fmt"
    "io/fs"
    "path/filepath"

    "github.com/yourusername/nodecleaner/pkg/models"
)

func (s *Scanner) walk(ctx context.Context, rootPath string) error {
    pathDepths := make(map[string]int)
    pathDepths[rootPath] = 0

    return filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
        // Check context cancellation
        select {
        case <-ctx.Done():
            return fs.SkipAll // Go 1.20+ - stops entire traversal
        default:
        }

        // Handle walk errors
        if err != nil {
            s.errors <- fmt.Errorf("accessing %s: %w", path, err)
            return fs.SkipDir // Skip this directory but continue walk
        }

        // Only process directories
        if !d.IsDir() {
            return nil
        }

        // Calculate current depth
        parentDepth := pathDepths[filepath.Dir(path)]
        currentDepth := parentDepth + 1
        pathDepths[path] = currentDepth

        // Respect max depth
        if s.config.MaxDepth > 0 && currentDepth > s.config.MaxDepth {
            return fs.SkipDir
        }

        // Check if path should be ignored
        if s.filter.ShouldIgnore(path) {
            return fs.SkipDir
        }

        // Check if this is a target directory (e.g., node_modules)
        if s.isTargetDir(d.Name()) {
            // Check cache first
            info, _ := d.Info()

            if s.cache != nil && s.cache.IsValid(path, info.ModTime()) {
                // Use cached data
                cached, _ := s.cache.Get(path)
                s.results <- models.DependencyFolder{
                    Path:       path,
                    Size:       cached.Size,
                    ModTime:    cached.ModTime,
                    AccessTime: cached.ModTime,
                    Type:       s.detectType(d.Name()),
                }
            } else {
                // Queue for analysis
                s.queueAnalysis(path)
            }

            // Don't recurse into target directories
            return fs.SkipDir
        }

        // Continue traversing
        return nil
    })
}

// Helper to queue directory for analysis
func (s *Scanner) queueAnalysis(path string) {
    // Send path to worker pool
    select {
    case s.workQueue <- path:
    default:
        // Queue full, analyze inline
        folder, err := s.analyzer.Analyze(path)
        if err != nil {
            s.errors <- fmt.Errorf("analyzing %s: %w", path, err)
            return
        }
        s.results <- *folder
    }
}

func (s *Scanner) isTargetDir(name string) bool {
    // Extensible for future dependency types
    targetDirs := []string{
        "node_modules", // JavaScript/TypeScript
        "vendor",       // Go/PHP
        ".venv",        // Python virtual environment
        "venv",         // Python virtual environment
        "target",       // Rust/Java
    }

    for _, target := range targetDirs {
        if name == target {
            return true
        }
    }
    return false
}

// detectType determines the dependency folder type
func (s *Scanner) detectType(name string) string {
    typeMap := map[string]string{
        "node_modules": "Node.js",
        "vendor":       "Go/PHP",
        ".venv":        "Python",
        "venv":         "Python",
        "target":       "Rust/Java",
    }

    if typ, ok := typeMap[name]; ok {
        return typ
    }
    return "Unknown"
}
```

### 6.2 Cache Invalidation Logic

**Strategy**: Timestamp-based with lazy pruning

```go
// internal/cache/invalidation.go

// ShouldRescan determines if a path needs rescanning
func (c *Cache) ShouldRescan(path string) (bool, error) {
    entry, exists := c.Get(path)
    if !exists {
        return true, nil // Not in cache
    }

    // Check if path still exists
    info, err := os.Stat(path)
    if os.IsNotExist(err) {
        // Path deleted, remove from cache
        c.mu.Lock()
        delete(c.index.Entries, path)
        c.modified = true
        c.mu.Unlock()
        return false, nil
    }
    if err != nil {
        return true, err
    }

    // Compare modification times
    if !info.ModTime().Equal(entry.ModTime) {
        return true, nil // Modified since last scan
    }

    // Check cache age (optional: invalidate after 7 days)
    cacheAge := time.Since(entry.LastScan)
    if cacheAge > 7*24*time.Hour {
        return true, nil
    }

    return false, nil
}
```

### 6.3 Size Calculation Optimization

**Strategy**: Parallel calculation with goroutines

```go
// internal/analyzer/stats.go

func (a *Analyzer) calculateSizeParallel(path string) (int64, error) {
    var (
        totalSize atomic.Int64
        wg        sync.WaitGroup
        errChan   = make(chan error, 1)
        semaphore = make(chan struct{}, 10) // Limit concurrent operations
    )

    err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil // Skip errors
        }

        if d.IsDir() {
            return nil
        }

        wg.Add(1)
        semaphore <- struct{}{} // Acquire

        go func(path string) {
            defer wg.Done()
            defer func() { <-semaphore }() // Release

            info, err := os.Stat(path)
            if err == nil {
                totalSize.Add(info.Size())
            }
        }(p)

        return nil
    })

    wg.Wait()
    close(errChan)

    if err != nil {
        return 0, err
    }

    return totalSize.Load(), nil
}
```

---

## 7. File System Operations

### 7.1 Path Filtering

```go
// internal/scanner/filter.go

type Filter struct {
    ignorePaths []string
    ignoreRegex []*regexp.Regexp
}

func NewFilter(ignorePaths []string) *Filter {
    return &Filter{
        ignorePaths: append(defaultIgnorePaths(), ignorePaths...),
    }
}

func defaultIgnorePaths() []string {
    return []string{
        "/System",
        "/Library",
        "/Applications",
        "/private/var",
        "/dev",
        "/proc",
        "/sys",
        "/.Trash",
        "/Network",
    }
}

func (f *Filter) ShouldIgnore(path string) bool {
    // Check exact matches
    for _, ignore := range f.ignorePaths {
        if strings.HasPrefix(path, ignore) {
            return true
        }
    }

    // Check hidden directories
    base := filepath.Base(path)
    if strings.HasPrefix(base, ".") && base != ".config" {
        return true
    }

    return false
}
```

### 7.2 Permission Handling

```go
// internal/scanner/permissions.go

func (s *Scanner) canAccess(path string) bool {
    // Try to stat the path
    _, err := os.Stat(path)
```
