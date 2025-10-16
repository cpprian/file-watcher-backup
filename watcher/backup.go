package watcher

// BackupManager handles creating and managing file backup with versioning.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cpprian/file-watcher-backup/utils"
)

type BackupManager struct {
	backupDir   string        // Directory where backup are stored
	maxVersions int           // Maximum number of versions to keep, the oldest are deleted
	logger      *utils.Logger // Logger instance for logging events
}

// NewBackupManager initializes a new BackupManager
func NewBackupManager(backupDir string, maxVersions int) *BackupManager {
	return &BackupManager{
		backupDir:   backupDir,
		maxVersions: maxVersions,
		logger:      utils.NewLogger(true, true),
	}
}

// CreateBackup creates a timestamped backup of the specified file
func (bm *BackupManager) CreateBackup(sourcePath, sourceDir string) error {
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	relPath, err := filepath.Rel(sourceDir, sourcePath)
	if err != nil {
		return fmt.Errorf("error while calculating relative path: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405.000000")

	ext := filepath.Ext(relPath)
	nameWithoutExt := strings.TrimSuffix(filepath.Base(relPath), ext)
	backupName := fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)

	fileVersionDir := filepath.Join(bm.backupDir, relPath+"_versions")
	backupPath := filepath.Join(fileVersionDir, backupName)

	if err := os.MkdirAll(fileVersionDir, 0755); err != nil {
		return fmt.Errorf("error while creating directory version: %w", err)
	}

	if err := utils.SafeCopyFile(sourcePath, backupPath, 3); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	bm.logger.BackupCreated(filepath.Base(sourcePath), backupName)

	if err := bm.cleanOldVersions(fileVersionDir, nameWithoutExt, ext); err != nil {
		return fmt.Errorf("error cleaning old versions: %w", err)
	}

	return nil
}

// cleanOldVersions remove old versions exceeding maxVersions
func (bm *BackupManager) cleanOldVersions(dir, baseName, ext string) error {
	pattern := filepath.Join(dir, fmt.Sprintf("%s_*%s", baseName, ext))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(matches) <= bm.maxVersions {
		return nil
	}

	sort.Strings(matches)

	toRemove := len(matches) - bm.maxVersions
	for i := range toRemove {
		if err := os.Remove(matches[i]); err != nil {
			return err
		}
		bm.logger.Info("	Removed old version: %s", filepath.Base(matches[i]))
	}

	return nil
}

// GetVersionCount returns the number of backup versions for a given file
func (bm *BackupManager) GetVersionCount(baseName, ext string) (int, error) {
	pattern := filepath.Join(bm.backupDir, fmt.Sprintf("%s_*%s", baseName, ext))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0, err
	}

	return len(matches), nil
}
