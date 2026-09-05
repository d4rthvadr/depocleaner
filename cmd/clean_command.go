package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/d4rthvadr/node-cleaner/internal/cache"
	"github.com/d4rthvadr/node-cleaner/internal/cleaner"
	"github.com/d4rthvadr/node-cleaner/internal/config"
	"github.com/d4rthvadr/node-cleaner/internal/logger"
	"github.com/d4rthvadr/node-cleaner/internal/scanner"
	"github.com/d4rthvadr/node-cleaner/internal/ui"
	"github.com/spf13/cobra"
)

var (
	noCacheClean bool
	dryRun       bool
	cleanPath    string
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Interactive clean (scan + select + delete)",
	RunE:  runClean,
}

func init() {

	cleanCmd.Flags().BoolVar(&noCacheClean, "no-cache", false, "Disable cache")
	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a preview run with no files deleted")
	cleanCmd.Flags().StringVar(&cleanPath, "path", "", "path to scan (default: $HOME)")

	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {

	ctx := cmd.Context()
	path := cleanPath

	cfg := config.Load()
	fmt.Printf("Scanning path: %s %v\n", path, cfg.ScanPath)

	if len(args) > 0 {
		path = args[0]
	}

	if path == "" {
		path = cfg.ScanPath // Use config default
	}
	cfg.ScanPath = path

	var c *cache.Cache
	var err error

	if !noCacheClean {
		c, err = cache.NewCache(cfg.CachePath)
		if err != nil {
			return err
		}
		defer c.Save()
	}

	scanner := scanner.NewScanner(cfg, c)

	result, err := scanner.Scan(ctx, path)
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	if len(result.Folders) == 0 {
		fmt.Println("No dependency folders found to clean.")
		return nil
	}

	// Interactive selection and deletion
	model := ui.NewSelectionModel(result.Folders)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running UI: %w", err)
	}

	selected := finalModel.(*ui.SelectionModel).GetSelectedFolders()

	if len(selected) == 0 {
		fmt.Println("No folders selected for deletion.")
		return nil
	}

	if !dryRun {
		fmt.Printf("\nAre you sure you want to delete %d selected folders? (y/n): ", len(selected))
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Aborting deletion.")
			return nil
		}
	}

	// Initialize logger
	log, err := logger.New(cfg.LogPath, logger.DefaultLogLevel)
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer log.Close()

	// Wrap logger with adapter for cleaner component
	cleanerLogger := logger.NewAdapter(log)
	cl := cleaner.NewCleaner(dryRun, cleanerLogger)

	cleanResult, err := cl.Clean(ctx, selected)

	if err != nil {
		return fmt.Errorf("cleaning folders: %w", err)
	}

	ui.DisplayCleanResults(cleanResult)
	return err
}
