package utils

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type BackupError struct {
	FilePath  string
	Operation string
	Err       error
	Retryable bool
}

func (e *BackupError) Error() string {
	return fmt.Sprintf("backup error [%s] %s: %v", e.Operation, e.FilePath, e.Err)
}

func IsRetryable(err error) bool {
	var backupErr *BackupError
	if errors.As(err, &backupErr) {
		return backupErr.Retryable
	}

	if errors.Is(err, os.ErrPermission) {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return false
}

func RetryWithBackoff(maxRetries int, initialDelay time.Duration, fn func() error) error {
	var lastErr error
	delay := initialDelay

	for i := range maxRetries {
		if err := fn(); err != nil {
			lastErr = err

			if !IsRetryable(err) {
				return err
			}

			if i < maxRetries-1 {
				time.Sleep(delay)
				delay *= 2
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("exceed max retries (%d): %w", maxRetries, lastErr)
}

func SafeCopyFile(src, dst string, maxRetries int) error {
	return RetryWithBackoff(maxRetries, 100*time.Millisecond, func() error {
		srcInfo, err := os.Stat(src)
		if err != nil {
			return &BackupError{
				FilePath:  src,
				Operation: "stat_source",
				Err:       err,
				Retryable: errors.Is(err, os.ErrPermission),
			}
		}

		if srcInfo.IsDir() {
			return &BackupError{
				FilePath:  src,
				Operation: "check_type",
				Err:       errors.New("source is a directory"),
				Retryable: false,
			}
		}

		srcFile, err := os.Open(src)
		if err != nil {
			return &BackupError{
				FilePath:  src,
				Operation: "open_source",
				Err:       err,
				Retryable: errors.Is(err, os.ErrPermission),
			}
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dst)
		if err != nil {
			return &BackupError{
				FilePath:  dst,
				Operation: "create_destination",
				Err:       err,
				Retryable: errors.Is(err, os.ErrPermission),
			}
		}
		defer dstFile.Close()

		buf := make([]byte, 32*1024)
		for {
			n, err := srcFile.Read(buf)
			if n > 0 {
				if _, err := dstFile.Write(buf[:n]); err != nil {
					return &BackupError{
						FilePath:  dst,
						Operation: "write",
						Err:       err,
						Retryable: true,
					}
				}
			}

			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				return &BackupError{
					FilePath: src,
					Operation: "read",
					Err: err,
					Retryable: true,
				}
			}
		}

		if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
			return nil
		}

		return nil
	})
}

func HandlePanic(logger *Logger, context string) {
	if r := recover(); r != nil {
		logger.Error("PANIC in %s: %v", context, r)
	}
}