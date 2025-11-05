# Boxify

A lightweight container runtime built from scratch in Go, similar to Docker but simpler. Boxify provides process isolation using Linux namespaces, resource management with cgroups, and networking capabilities.

## Features

- **Process Isolation**: Uses Linux namespaces (PID, UTS, IPC, NET, MNS) for container isolation
- **Resource Management**: CPU and memory limits via cgroups v2
- **Networking**: Virtual ethernet pairs with bridge networking
- **Overlay Filesystem**: Uses overlay mounts for container filesystem isolation
- **Daemon Architecture**: Background daemon manages container lifecycle
- **Interactive Shell**: Attach to running containers with `nsenter`

## Architecture

```
boxify (client) --> HTTP over Unix Socket --> boxifyd (daemon) --> boxify-init (container process)
                                                                          |
                                                                          v
                                                                    Container Namespaces
                                                                    - PID, UTS, IPC, NET, MNS
                                                                    - Overlay FS
                                                                    - Network setup
```

## Prerequisites

- **OS**: Linux (tested on Arch Linux with kernel 6.17+)
- **Root Access**: Required for namespace operations, cgroups, and networking
- **Go**: 1.21 or later
- **Dependencies**:
  - `nsenter` utility (part of `util-linux`)
  - `systemd` (for daemon management)

## Quick Start

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd boxify

# Run the setup (requires root)
sudo make setup
```

That's it! The `make setup` command will:
1. Stop any running boxifyd daemon
2. Extract the included Alpine rootfs to `/var/lib/boxify/boxify-rootfs`
3. Build all binaries (`boxify`, `boxifyd`, `boxify-init`)
4. Install binaries to `/usr/local/bin/`
5. Install systemd service
6. Start the boxifyd daemon

### Running Your First Container

```bash
# Create a config file from the example
cp boxify.example.yaml boxify.yaml

# Run the container
sudo boxify run
```

You'll be dropped into an interactive Alpine Linux shell inside the container!

## Configuration

Create a `boxify.yaml` file in your working directory:

```yaml
image_name: nodejs
settings:
  memory_limit: 100m
  cpu_limit: 2
```

See `boxify.example.yaml` for reference.

**Configuration Options:**
- `image_name`: Name for your container (used for identification)
- `memory_limit`: Maximum memory (e.g., `100m`, `1g`)
- `cpu_limit`: CPU weight (relative CPU time, higher = more CPU)

## Usage

### Running a Container

```bash
sudo boxify run
```

This will:
1. Read `boxify.yaml` configuration
2. Send a request to the daemon to create a container
3. Receive the container PID
4. Automatically attach to the container using `nsenter`
5. Drop you into an interactive shell inside the container

### Inside the Container

Once attached, you're in an isolated Alpine Linux environment:

```bash
# Check the hostname (should be different from host)
hostname

# List processes (only container processes visible)
ps aux

# Check network configuration
ip addr

# Test network connectivity
ping -c 3 8.8.8.8

# Install packages
apk add curl vim

# Exit the container
exit
```

### Managing the Daemon

```bash
# Check daemon status
sudo systemctl status boxifyd

# View daemon logs
sudo journalctl -u boxifyd -f

# Restart daemon
sudo systemctl restart boxifyd

# Stop daemon
sudo systemctl stop boxifyd
```

## How It Works

### Container Creation Flow

1. **Client** (`boxify`):
   - Reads `boxify.yaml` configuration
   - Sends HTTP request to daemon via Unix socket at `/var/run/boxify.sock`

2. **Daemon** (`boxifyd`):
   - Creates network infrastructure (veth pair, bridge)
   - Generates unique container ID (UUID)
   - Creates overlay filesystem from Alpine rootfs
   - Spawns `boxify-init` with new namespaces (CLONE_NEWUTS, CLONE_NEWPID, CLONE_NEWIPC, CLONE_NEWNET, CLONE_NEWNS)
   - Sets up cgroups for resource limits
   - Moves veth into container's network namespace
   - Returns container PID to client

3. **Container Init** (`boxify-init`):
   - Performs `chroot` into overlay filesystem
   - Mounts `/proc`, `/sys`, `/dev`
   - Configures container networking (IP address, routes)
   - Blocks indefinitely (waiting for attach)

4. **Client Attach**:
   - Uses `nsenter` to enter container's namespaces
   - Spawns `/bin/sh` inside the container
   - Provides interactive shell to user

### Filesystem Isolation

Boxify uses overlay filesystem with the following structure:

```
/var/lib/boxify/boxify-container/<containerID>/
├── lower/          # Symlink to Alpine rootfs (read-only)
├── upper/          # Container-specific changes (read-write)
├── work/           # Overlay work directory
└── merged/         # Combined view (what container sees)
```

The Alpine rootfs is extracted to `/var/lib/boxify/boxify-rootfs/` and used as the lower layer for all containers.

### Networking

- **Bridge**: `boxify-bridge0` (10.88.0.0/16)
- **Container IPs**: Auto-assigned from bridge subnet (starting at 10.88.0.12)
- **veth Pairs**: Virtual ethernet pairs connect containers to bridge
- **Gateway**: 10.88.0.0
- **NAT**: Traffic routed through host

## Makefile Commands

```bash
# Setup everything (extract rootfs, build, install, start daemon)
sudo make setup

# Clean up rootfs
sudo make clean

# Build and run (for development)
make run
```

## Troubleshooting

### Container exits immediately

**Symptom**: `nsenter: reassociate to namespaces failed: No such process`

**Cause**: The `boxify-init` process is exiting before `nsenter` can attach.

**Solution**: Rebuild `boxify-init` with the latest code:
```bash
go build -o boxify-init ./cmd/boxify-init
sudo cp boxify-init /usr/local/bin/boxify-init
sudo chmod +x /usr/local/bin/boxify-init
```

### Permission denied errors

**Solution**: Boxify requires root privileges. Always run with `sudo`.

### Network not working inside container

**Check**:
```bash
# On host, verify bridge exists
ip link show boxify-bridge0

# Check daemon logs
sudo journalctl -u boxifyd -f
```

### Daemon won't start

**Check logs**:
```bash
sudo journalctl -u boxifyd -n 50
```

Common issues:
- Socket already in use (stop old daemon: `sudo systemctl stop boxifyd`)
- Missing directories (run `sudo make setup`)
- Insufficient permissions (must run as root)

### Container filesystem is empty

**Solution**: Run `sudo make setup` to extract the Alpine rootfs.

### Config file not found

**Error**: `Config file not found: boxify.yaml`

**Solution**: Create a `boxify.yaml` file in the current directory or copy from example:
```bash
cp boxify.example.yaml boxify.yaml
```

## Cleanup

### Remove all container data

```bash
sudo rm -rf /var/lib/boxify/boxify-container/*
```

### Uninstall completely

```bash
# Stop the daemon
sudo systemctl stop boxifyd
sudo systemctl disable boxifyd

# Remove binaries
sudo rm /usr/local/bin/boxify
sudo rm /usr/local/bin/boxifyd
sudo rm /usr/local/bin/boxify-init

# Remove systemd service
sudo rm /etc/systemd/system/boxifyd.service
sudo systemctl daemon-reload

# Remove data directories
sudo rm -rf /var/lib/boxify
```

## Project Structure

```
boxify/
├── cmd/
│   ├── boxify/              # Client CLI
│   │   └── main.go
│   ├── boxify-init/         # Container init process
│   │   └── main.go
│   └── boxifyd/             # Daemon
│       └── main.go
├── pkg/
│   ├── cgroup/              # Cgroups v2 management
│   ├── container/           # Container/overlay filesystem
│   ├── daemon/              # Daemon handlers and types
│   │   ├── boxifyd.service  # Systemd service file
│   │   ├── handlers/        # HTTP request handlers
│   │   ├── requests/        # Request types
│   │   └── types/           # Container types
│   └── network/             # Networking (bridge, veth, IP management)
├── config/                  # Configuration structures
├── alpine-minirootfs-*.tar.gz  # Alpine Linux rootfs (included)
├── boxify.example.yaml      # Example configuration
├── Makefile                 # Build and setup automation
└── README.md                # This file
```

## Limitations

- Single container per `boxify run` command (containers are ephemeral)
- No image management (uses Alpine rootfs directly)
- No container persistence between runs
- No volume mounts
- No port forwarding configuration
- No container listing/management commands
- Limited to Linux systems with cgroups v2
- Containers cleanup automatically on exit

## Development

### Building from source

```bash
# Build all components
go build -o boxify ./cmd/boxify
go build -o boxifyd ./cmd/boxifyd
go build -o boxify-init ./cmd/boxify-init

# Install to system
sudo cp boxify boxifyd boxify-init /usr/local/bin/
sudo chmod +x /usr/local/bin/boxify*
```

### Testing

```bash
# Quick test
make run
```

### Viewing logs

```bash
# Follow daemon logs
sudo journalctl -u boxifyd -f

# View recent logs
sudo journalctl -u boxifyd -n 100
```

## Contributing

This is an educational project demonstrating container runtime fundamentals. Contributions, issues, and feature requests are welcome!

## License

[Add your license here]

## References

- [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- [Control Groups v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)
- [Overlay Filesystem](https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html)
- [Alpine Linux](https://alpinelinux.org/)
