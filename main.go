package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cpprian/file-watcher-backup/config"
	"github.com/cpprian/file-watcher-backup/utils"
	"github.com/cpprian/file-watcher-backup/watcher"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App {
		Name: "file-watcher-backup",
		Usage: "Monitors a directory and creates backups of changed files.",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "source",
				Aliases:  []string{"s"},
				Usage:    "Directory to monitor for changes",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "backup",
				Aliases:  []string{"b"},
				Usage:    "Directory to store backups",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "versions",
				Aliases: []string{"vers"},
				Usage:   "Maximum number of versions to store per file",
				Value:   3,
			},
			&cli.DurationFlag{
				Name:    "interval",
				Aliases: []string{"i"},
				Usage:   "Interval between scans for changes",
				Value:   5 * time.Second,
			},
		},
		Action: runWatcher,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runWatcher(c *cli.Context) error {
	startTime := time.Now()
	logger := utils.NewLogger(true, true)

	source := c.String("source")
	backup := c.String("backup")
	versions := c.Int("versions")
	interval := c.Duration("interval")

	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", source)
	}

	if err := os.MkdirAll(backup, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	cfg := config.NewConfig(source, backup, versions, interval)

	fw, err := watcher.NewFileWatcher(cfg)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- fw.Start()
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			ticker.Stop()
			fw.Stop()

			duration := time.Since(startTime)
			logger.ShutdownComplete(duration)
		
		case err := <-errChan:
			return fmt.Errorf("error watcher: %w", err)

		case <-ticker.C:
			stats := fw.GetStats()
			logger.Stats(
				stats["tracked_files"].(int),
				stats["queue_length"].(int),
				stats["queue_capacity"].(int),
				stats["active_workers"].(int),
			)
		}
	}
}