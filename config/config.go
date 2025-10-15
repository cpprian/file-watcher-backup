package config

// Config holds the configuration settings for the backup tool

import "time"

type Config struct {
	SourceDir      string        // Directory to monitor
	BackupDir      string        // Directory to store backups
	MaxVersions    int           // Maximum number of backup versions to keep
	MinInterval    time.Duration // Minimum interval between backups
	IgnorePatterns []string      // Patterns to ignore when monitoring files
}

// TODO: In the future, this could be loaded from a file
// NewConfig creates a new Config instance with default ignore patterns 
func NewConfig(source, backup string, versions int, interval time.Duration) *Config {
	return &Config{
		SourceDir:   source,
		BackupDir:   backup,
		MaxVersions: versions,
		MinInterval: interval,
		IgnorePatterns: []string{
			"*.tmp",
			"*.swp",
			".git",
			".DS_Store",
		},
	}
}
