# garbage

Soft-delete and trash management CLI for the Amadla ecosystem.

Files and directories are moved to a restorable trash directory (`~/.local/amadla/trash/`) with metadata tracked in a shared SQLite database (`~/.local/amadla/db.sqlite`).

## Installation

```bash
make build
```

Requires CGO (for SQLite via `mattn/go-sqlite3`).

## Usage

### Trash a file or directory

```bash
garbage rm /path/to/file.txt
garbage rm /path/to/directory /path/to/another-file.txt
```

### List trashed items

```bash
garbage list
garbage list --type=file
garbage list --type=directory
```

### Show item details

```bash
garbage info <id>
```

### Restore a trashed item

```bash
garbage restore <id>           # Restore to original location
garbage restore <id> --to=/new/path
garbage restore myfile.txt     # Restore by name
```

### Permanently delete trashed items

```bash
garbage empty --force                # Delete all
garbage empty --older-than=30d --force  # Delete items older than 30 days
```

### Show settings and statistics

```bash
garbage settings
```

### Global flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `--quiet` | Suppress non-error output |
| `--verbose` | Show detailed output |
| `--dry-run` | Show what would be done without making changes |

## Development

```bash
make test           # Run tests
make lint           # Run linter
make build          # Build for current platform
make build-all      # Build for all platforms
```
