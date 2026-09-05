package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/d4rthvadr/node-cleaner/internal/analyzer"
	"github.com/d4rthvadr/node-cleaner/pkg/models"
	"github.com/d4rthvadr/node-cleaner/pkg/utils"
)

type Scanner struct {
	config    *models.Config
	cache     CacheProvider
	results   chan models.DependencyFolder
	errorChan chan error
	analyzer  *analyzer.Analyzer
	workQueue chan string
}

type CacheProvider interface {
	Get(path string) (*models.CacheEntry, bool)
	Set(path string, entry *models.CacheEntry) error
	IsValid(path string, modTime time.Time) bool
	Save() error
}

// NewScanner creates a new Scanner instance
func NewScanner(cfg *models.Config, cache CacheProvider) *Scanner {
	return &Scanner{
		analyzer:  analyzer.NewAnalyzer(),
		config:    cfg,
		cache:     cache,
		results:   make(chan models.DependencyFolder, 100), // buffered to prevent blocking
		errorChan: make(chan error, 50),                    // buffered for errors
	}
}

// Scan initiates file traversal process
func (s *Scanner) Scan(ctx context.Context, rootPath string) (*models.ScanResult, error) {

	finalResult := &models.ScanResult{
		ScanPath: rootPath,
		ScanTime: time.Now(),
	}

	fmt.Println("Starting scan on path:", rootPath)

	s.workQueue = make(chan string, s.config.Workers*2) // buffered channel

	var wg sync.WaitGroup

	// start workers
	for i := 0; i < s.config.Workers; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg)
	}

	// walk the file system starting from rootPath
	// and send directories to be processed by workers

	go func() {
		if err := s.walkFileSystem(ctx, rootPath, 0); err != nil {
			s.errorChan <- fmt.Errorf("walking filesystem: %w", err)
			s.handleError(fmt.Errorf("error occured %v", err), rootPath)

		}
		fmt.Println("file system walk completed")
		close(s.workQueue)
	}()

	go func() {
		wg.Wait()
		close(s.results)
		close(s.errorChan)
	}()

	done := make(chan struct{})

	go func() {
		for err := range s.errorChan {
			// Log errors (could be aggregated or handled differently)
			fmt.Printf("Scan error: %v\n", err)
		}
		done <- struct{}{}
	}()

	// aggregate results and errors concurrently

	for r := range s.results {
		finalResult.Folders = append(finalResult.Folders, r)
		finalResult.TotalSize += r.Size
		finalResult.TotalCount++
	}

	<-done // wait for error processing to complete

	finalResult.Duration = time.Since(finalResult.ScanTime)
	return finalResult, nil

}

func (s *Scanner) walkFileSystem(ctx context.Context, rootPath string, depth int) error {

	pathDepths := make(map[string]int)
	pathDepths[rootPath] = depth
	// WalkDir is a convenient way to traverse directories tho
	// it may need to be replaced with a custom implementation
	// if we need to follow symlinks
	return filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {

		// check for context cancellation
		select {
		case <-ctx.Done():
			return fs.SkipAll
		default:
			// continue processing
		}

		if err != nil {
			s.errorChan <- fmt.Errorf("accessing %s: %w", path, err)
			return fs.SkipDir // skip this directory on error but keep walking
		}

		if !d.IsDir() {
			return nil // we only care about directories
		}

		parentDepth := pathDepths[filepath.Dir(path)]
		currentDepth := parentDepth + 1
		pathDepths[path] = currentDepth

		// check max depth
		if s.config.MaxDepth > 0 && currentDepth > s.config.MaxDepth {
			return fs.SkipDir
		}

		// TODO: check ignore paths

		if utils.IsTargetDirectory(d.Name()) {

			info, _ := d.Info()

			if s.cache != nil && s.cache.IsValid(path, info.ModTime()) {

				// use cached data
				cached, cacheHit := s.cache.Get(path)
				fmt.Println("cache hit? ", cacheHit)
				s.results <- models.DependencyFolder{
					Path:         path,
					AbsolutePath: path,
					Size:         cached.Size,
					ModTime:      cached.ModTime,
					Type:         utils.DetectType(d.Name()),
				}

			} else {
				s.enqueueAndAnalysis(path)
			}

			return fs.SkipDir // skip further traversal into this directory

		}

		// Continue walking
		return nil

	})

}

func (s *Scanner) enqueueAndAnalysis(path string) {

	// send path to worker pool
	select {
	case s.workQueue <- path:
	default:
		// if workQueue is full(aka workers are busy), process immediately here
		// so that path wont be lost

		folder, err := s.analyzer.Analyze(path)
		if err != nil {
			s.handleError(fmt.Errorf("analyzing %s: %w", path, err), path)
			return
		}
		s.results <- *folder
	}

}

// worker processes directories and sends results/errors

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

			folder, err := s.analyzer.Analyze(path)
			if err != nil {
				s.handleError(fmt.Errorf("analyzing %s: %w", path, err), path)
				continue
			}

			// Cache the result if caching is enabled
			if s.cache != nil {
				s.cache.Set(path, &models.CacheEntry{
					Path:     path,
					Size:     folder.Size,
					ModTime:  folder.ModTime,
					LastScan: time.Now(),
				})
				err := s.cache.Save()
				if err != nil {
					fmt.Printf("failed to save cache: %v\n", err)
				}
			}

			// context could cancelled while sending result
			select {
			case s.results <- *folder:
			case <-ctx.Done():
				return

			}

		}
	}
}
