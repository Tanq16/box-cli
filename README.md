<div align="center">
  <img src=".github/assets/logo.png" alt="Box Logo" width="200">
  <h1>Box</h1>

  <a href="https://github.com/tanq16/box/actions/workflows/release.yml"><img alt="Build Workflow" src="https://github.com/tanq16/box/actions/workflows/release.yml/badge.svg"></a>&nbsp;<a href="https://github.com/tanq16/box/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/tanq16/box"></a><br><br>
  <a href="#capabilities">Capabilities</a> &bull; <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#tips-and-notes">Tips & Notes</a>
</div>

---

A CLI tool for Box.com file operations. Single binary, cross-platform, built with Go.

## Capabilities

| Category | Commands | Description |
|----------|----------|-------------|
| Auth | `login` | OAuth authentication with Box |
| Files | `list`, `upload`, `download`, `delete` | Browse, upload, download, and delete files/folders |
| Manage | `mkdir`, `move`, `copy`, `info` | Create folders, move/rename, copy, and inspect items |
| Search | `search` | Server-side search with filters |
| Sync | `sync push`, `sync pull` | Bidirectional sync between local and Box |
| Sharing | `shared-link create/get/remove/resolve` | Manage shared links on items |
| Collaboration | `collab create/get/update/delete/pending` | Manage collaborations |

## Installation

### Binary

Download from [releases](https://github.com/tanq16/box/releases):

```bash
# Linux/macOS
curl -sL https://github.com/tanq16/box/releases/latest/download/box-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o box
chmod +x box
sudo mv box /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/tanq16/box
cd box
make build
```

## Usage

### Setup

1. Create a Box application at [developer.box.com](https://developer.box.com) with OAuth 2.0 authentication
2. Save your credentials:

```bash
mkdir -p ~/.box
cat > ~/.box/credentials.json <<EOF
{
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET"
}
EOF
```

3. Authenticate:

```bash
box login
```

### File Operations

#### `list`

List contents of a Box folder. Use `--filter` / `-F` for case-insensitive name filtering.

```bash
box list /path/to/folder
box list --id FOLDER_ID
box list / -F reports      # Only show items with "reports" in the name
```

#### `upload`

Upload a file or folder to Box. Files over 50MB automatically use chunked upload.

```bash
box upload ./local-file.txt /remote/path
box upload ./local-folder /remote/path
box upload large-file.zip /remote --chunked
```

#### `download`

Download a file or folder from Box.

```bash
box download /remote/file.txt ./local-path
box download /remote/folder ./local-folder
box download --id FILE_ID
```

#### `delete`

Delete a file or folder on Box.

```bash
box delete /path/to/file.txt
box delete /path/to/folder
box delete --id ITEM_ID
```

#### `mkdir`

Create a folder on Box. Use `-p` to create intermediate directories.

```bash
box mkdir /Documents/new-folder
box mkdir -p /deep/nested/path/new-folder
```

#### `move`

Move or rename a file or folder. If the destination is an existing folder, the item is moved into it. If the source and destination share the same parent, it acts as a rename.

```bash
box move /old/path/file.txt /new/location/       # Move into folder
box move /path/old-name.txt /path/new-name.txt    # Rename in place
box move /folder-a/doc.txt /folder-b/renamed.txt  # Move and rename
```

#### `copy`

Copy a file or folder. Use `--name` to give the copy a different name.

```bash
box copy /source/file.txt /dest/folder/
box copy /source/folder /dest/folder/
box copy /source/file.txt /dest/folder/new-name.txt
box copy /source/file.txt /dest/folder/ --name copy-of-file.txt
```

#### `info`

Show metadata for a file or folder.

```bash
box info /path/to/file.txt
box info /path/to/folder
box info --id ITEM_ID
```

### Search

#### `search`

Server-side search with filters. Time filters use relative shorthands: `s` (seconds), `m` (minutes), `h` (hours), `d` (days), `w` (weeks), `M` (months), `y` (years).

```bash
box search "quarterly report"
box search "budget" --type file --extensions xlsx,csv --limit 10
box search "notes" --folder-id 12345 --sort modified_at
box search "report" --created-in 2w    # Created in the last 2 weeks
box search "data" --updated-in 3M     # Updated in the last 3 months
```

### Sync

#### `sync push` / `sync pull`

Bidirectional sync with hash-based change detection.

```bash
box sync push ./local-dir /remote/dir --concurrency 8
box sync pull /remote/dir ./local-dir --ignore ".git,.DS_Store"
```

### Shared Links

Manage shared links on files and folders. Supports both path and `--id` resolution.

```bash
box shared-link create /path/to/file --access open
box shared-link create --id ITEM_ID -a company -P secretpass
box shared-link get /path/to/folder
box shared-link remove /path/to/file
box shared-link resolve "https://app.box.com/s/..."
```

### Collaborations

Share files and folders with users by email. The `create` subcommand takes a path and email as positional args and defaults to `viewer` role.

```bash
box collab create /path/to/folder user@example.com
box collab create /path/to/file user@example.com -r editor
box collab create --id ITEM_ID user@example.com -r co-owner
box collab get COLLAB_ID
box collab update COLLAB_ID -r editor
box collab delete COLLAB_ID
box collab pending
```

## Tips and Notes

- Use `--debug` flag on any command for detailed logging output
- Files over 50MB are automatically uploaded via chunked upload
- Sync uses SHA1 hashes to detect changes efficiently
- Use `--concurrency` flag with sync commands to control parallelism (default: 4)
