package watcher

// FileWatcher implements a file system watcher that monitors a specified directory
// and its subdirectories for file changes.
//
// It uses fsnotify to watch for events such as file creation, modification,
// deletion, and renaming. When a file is created or modified, it enqueues a
// backup job to be processed by a pool of worker goroutines.
//
// The FileWatcher also respects ignore patterns and ensures
// that backups are not created too frequently for the same file.

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cpprian/file-watcher-backup/config"
	"github.com/cpprian/file-watcher-backup/utils"
	"github.com/fsnotify/fsnotify"
)

// BackupJob represents a job to back up a specific file
type BackupJob struct {
	FilePath  string    // Absolute path to the file
	EventType string    // Type of event (e.g., "CREATE", "MODIFY")
	Timestamp time.Time // Time when the event was detected
}

// FileWatcher monitors file system events and manages backup jobs
type FileWatcher struct {
	config        *config.Config       // Configuration settings
	BackupManager *BackupManager       // Manages backup operations
	watcher       *fsnotify.Watcher    // fsnotify watcher instance
	lastBackup    map[string]time.Time // Tracks last backup times for files
	mu            sync.Mutex           // Mutex for synchronizing access to lastBackup
	backupQueue   chan BackupJob       // Channel for backup jobs
	workerWg      sync.WaitGroup       // WaitGroup for worker goroutines
	stopChan      chan struct{}        // Channel to signal stopping the watcher
	numWorkers    int                  // Number of worker goroutines
	logger        *utils.Logger        // Logger for logging events and errors
}

// NewFileWatcher creates a new FileWatcher instance with the provided configuration
func NewFileWatcher(cfg *config.Config) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %w", err)
	}

	return &FileWatcher{
		config:        cfg,
		BackupManager: NewBackupManager(cfg.BackupDir, cfg.MaxVersions),
		watcher:       watcher,
		lastBackup:    make(map[string]time.Time),
		backupQueue:   make(chan BackupJob, 100),
		stopChan:      make(chan struct{}),
		numWorkers:    3,
		logger:        utils.NewLogger(true, true),
	}, nil
}

// Start begins watching the configured directory for file changes
func (fw *FileWatcher) Start() error {
	if err := fw.addDirectoryRecursive(fw.config.SourceDir); err != nil {
		return fmt.Errorf("error adding directory: %w", err)
	}

	fw.logger.Headder(
		fw.config.SourceDir,
		fw.config.BackupDir,
		fw.config.MaxVersions,
		fw.numWorkers,
	)

	fw.startWorkerPool()

	go fw.watchLoop()

	<-fw.stopChan
	return nil
}

// startWorkerPool initializes the pool of worker goroutines
func (fw *FileWatcher) startWorkerPool() {
	for i := range fw.numWorkers {
		fw.workerWg.Add(1)
		go fw.backupWorker(i + 1)
	}
}

// backupWorker processes backup jobs from the queue
func (fw *FileWatcher) backupWorker(id int) {
	defer fw.workerWg.Done()
	defer utils.HandlePanic(fw.logger, fmt.Sprintf("Worker #%d", id))

	for job := range fw.backupQueue {
		fw.logger.WorkerStarted(id, filepath.Base(job.FilePath))

		if err := fw.BackupManager.CreateBackup(job.FilePath, fw.config.SourceDir); err != nil {
			fw.logger.Error("Worker #%d: %v", id, err)
		}
	}
}

// watchLoop continuously listens for file system events and errors
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}

			log.Printf("❌ Error from watcher: %v\n", err)
		}
	}
}

// hanldeEvent processes a single fsnotify event
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	var eventType string

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = "CREATe"

		if isDir(event.Name) {
			fw.watcher.Add(event.Name)
			fw.logger.Info("New catalog: %s", filepath.Base(event.Name))
		}
		fw.logger.FileCreated(filepath.Base(event.Name))

	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = "WRITE"
		fw.logger.FileModified(filepath.Base(event.Name))

	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = "REMOVE"
		fw.logger.FileDeleted(filepath.Base(event.Name))

		// While removing there is no any sense to backup
		return

	case event.Op&fsnotify.Rename == fsnotify.Rename:
		eventType = "RENAME"
		fw.logger.FileRenamed(filepath.Base(event.Name))
		return

	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		// TODO: Handle chmod if necessary
		return

	default:
		return
	}

	if isDir(event.Name) {
		return
	}

	if fw.shouldIgnore(event.Name) {
		return
	}

	fw.enqueueBackup(event.Name, eventType)
}

// enqueueBackup adds a backup job to the queue if conditions are met
func (fw *FileWatcher) enqueueBackup(path string, eventType string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	lastTime, exists := fw.lastBackup[path]
	if exists && time.Since(lastTime) < fw.config.MinInterval {
		fw.logger.BackupSkipped(filepath.Base(path), "too soon since last backup")
		return
	}

	job := BackupJob{
		FilePath:  path,
		EventType: eventType,
		Timestamp: time.Now(),
	}

	select {
	case fw.backupQueue <- job:
		fw.lastBackup[path] = time.Now()
		fw.logger.Info("Add to backup queue: %s [%s]", filepath.Base(path), eventType)

	default:
		fw.logger.Warning("Queue full, skipping backup for: %s", filepath.Base(path))
	}
}

// addDirectoryRecursive adds a directory and its subdirectories to the watcher
func (fw *FileWatcher) addDirectoryRecursive(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fw.shouldIgnore(walkPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if isDir(walkPath) {
			if err := fw.watcher.Add(walkPath); err != nil {
				return err
			}
		}

		return nil
	})
}

// shouldIgnore checks if a file or directory should be ignored based on the ignore patterns
func (fw *FileWatcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range fw.config.IgnorePatterns {
		matched, _ := filepath.Match(pattern, base)
		if matched {
			return true
		}

		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// isDir checks if the given path is a directory
func isDir(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetStats returns statistics about the FileWatcher
func (fw *FileWatcher) GetStats() map[string]interface{} {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	return map[string]interface{}{
		"tracked_files":  len(fw.lastBackup),
		"queue_length":   len(fw.backupQueue),
		"queue_capacity": cap(fw.backupQueue),
		"active_workers": fw.numWorkers,
	}
}

// Stop gracefully stops the FileWatcher and all its workers
func (fw *FileWatcher) Stop() {
	fw.logger.Shutdown()

	close(fw.backupQueue)

	fw.workerWg.Wait()

	fw.watcher.Close()

	close(fw.stopChan)

	fw.logger.Success("Watcher stopped")
}

// Close implements the io.Closer interface for FileWatcher
func (fw *FileWatcher) Close() error {
	fw.Stop()
	return nil
}
