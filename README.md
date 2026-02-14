# box

A CLI tool for Box.com file operations. Single binary, cross-platform, built with Go.

## Install

Download a binary from the [releases page](https://github.com/tanq16/box/releases), or build from source:

```bash
git clone https://github.com/tanq16/box.git
cd box
make build
```

## Setup

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

## Commands

| Command | Description |
|---------|-------------|
| `box login` | Authenticate with Box via OAuth |
| `box list [path]` | List contents of a folder |
| `box upload <local> [remote]` | Upload a file or folder |
| `box download <remote> [local]` | Download a file or folder |
| `box sync push <local> <remote>` | Sync local directory to Box |
| `box sync pull <remote> <local>` | Sync Box directory to local |
| `box search <query>` | Search files on Box (server-side) |
| `box index [path]` | Build local index of Box contents |
| `box index search <regex>` | Search local index |
| `box shared-link create <path>` | Create a shared link |
| `box shared-link get <path>` | Get shared link info |
| `box shared-link remove <path>` | Remove a shared link |
| `box shared-link resolve <url>` | Resolve a shared link URL |
| `box collab create` | Create a collaboration |
| `box collab get <id>` | Get collaboration details |
| `box collab update <id>` | Update a collaboration |
| `box collab delete <id>` | Delete a collaboration |
| `box collab pending` | List pending collaborations |

## Build

```bash
make build          # Build for current platform
make build-all      # Build for all platforms (linux, darwin, windows; amd64, arm64)
make tidy           # Run go mod tidy
make clean          # Remove build artifacts
```
