# File Watcher & Auto-Backup CLI

A command-line tool that monitors specified files and creates backups whenever changes are detected with versioning support.

## Features

- Monitor multiple files for changes
- Timestamped backups with precise microsecond resolution
- Versioning support to keep track of multiple changes
- Miminal delay between backups to avoid excessive file creation
- Recursive directory monitoring
- Worker pool - process multiple files concurrently
- Ignoring specific files or directories (e.g., `.tmp`, `.DS_Store`, `.git`)
- Retry mechanism for robustness
- Color-coded terminal output for better readability

## Installation

### Prerequisites

- Go 1.21 of higher

### Steps

1. Clone the repository:

   ```bash
   git clone https://github.com/cyprian/file-watcher-backup
    cd file-watcher-backup
    ```

2. Build the project:

   ```bash
   go mod download
   go build -o file-watcher
   ```

3. (Optional) Move the binary to global path:

   ```bash
   mv file-watcher /usr/local/bin/
   ```

### Usage

```bash
./file-watcher --source ./my-project --backup ./backups
```

Or with all options:

```bash
./file-watcher \
  --source ./my-project \
  --backup ./backups \
  --versions 5 \
  --interval 10s
```

## Command-Line Options

- `--source` (string, required): Path to the source file or directory to monitor.
- `--backup` (string, required): Path to the backup directory where backups will be stored.
- `--versions` (int, default: 3): Number of backup versions to keep for each file.
- `--interval` (duration, default: 5s): Minimum interval between backups.

## Todo list

- [ ] Configure delay time
- [ ] Add tests
- [ ] Add command to load ignoring paths or files from a file or multiple arguments (e.g., `--ignore .tmp .DS_Store .git`)
- [ ] Add support for backup compression
- [ ] Add performance benchmarks