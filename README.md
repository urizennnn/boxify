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
  - Alpine Linux rootfs

## Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd boxify
```

### 2. Set Up Alpine Root Filesystem

```bash
# Create the rootfs directory
sudo mkdir -p /var/lib/boxify/boxify-rootfs

# Download Alpine mini root filesystem
wget https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.0-x86_64.tar.gz

# Extract to the rootfs directory
sudo tar -xzf alpine-minirootfs-3.19.0-x86_64.tar.gz -C /var/lib/boxify/boxify-rootfs/

# Clean up
rm alpine-minirootfs-3.19.0-x86_64.tar.gz
```

### 3. Build Boxify Components

```bash
# Build the daemon
sudo go build -o /usr/local/bin/boxifyd ./cmd/boxifyd

# Build the container init process
sudo go build -o /usr/local/bin/boxify-init ./cmd/boxify-init

# Build the client
sudo go build -o /usr/local/bin/boxify ./cmd/boxify
```

### 4. Create Required Directories

```bash
sudo mkdir -p /var/lib/boxify/networks
sudo mkdir -p /var/lib/boxify/boxify-container
```

### 5. Start the Daemon

**Option A: Run directly**
```bash
sudo /usr/local/bin/boxifyd
```

**Option B: Using systemd (recommended)**

Create `/etc/systemd/system/boxifyd.service`:
```ini
[Unit]
Description=Boxify Container Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/boxifyd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable boxifyd
sudo systemctl start boxifyd

# Check status
sudo systemctl status boxifyd
```

## Usage

### 1. Create Configuration File

Create a `boxify.yaml` (or `boxify.yml`) in your project directory:

```yaml
imageName: nodejs
settings:
  memoryLimit: 100m
  cpuLimit: 2
```

- **imageName**: Name for your container (used for identification)
- **memoryLimit**: Maximum memory (e.g., `100m`, `1g`)
- **cpuLimit**: CPU weight (relative CPU time)

### 2. Run a Container

```bash
sudo boxify run
```

This will:
1. Read the configuration file
2. Send a request to the daemon to create a container
3. Receive the container PID
4. Automatically attach to the container using `nsenter`
5. Drop you into an interactive shell inside the container

### 3. Inside the Container

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

# Install packages (if needed)
apk add curl

# Exit the container
exit
```

## How It Works

### Container Creation Flow

1. **Client** (`boxify`):
   - Reads `boxify.yaml` configuration
   - Sends HTTP request to daemon via Unix socket at `/var/run/boxify.sock`

2. **Daemon** (`boxifyd`):
   - Creates network infrastructure (veth pair, bridge)
   - Generates unique container ID
   - Creates overlay filesystem
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
├── lower/          # Alpine rootfs (read-only)
├── upper/          # Container-specific changes (read-write)
├── work/           # Overlay work directory
└── merged/         # Combined view (what container sees)
```

### Networking

- **Bridge**: `boxify-bridge0` (10.88.0.0/16)
- **Container IPs**: Auto-assigned from bridge subnet
- **veth Pairs**: Virtual ethernet pairs connect containers to bridge
- **NAT**: Traffic routed through host

## Troubleshooting

### Container exits immediately

**Symptom**: `nsenter: reassociate to namespaces failed: No such process`

**Solution**: Make sure `boxify-init` is built with the latest code that blocks forever instead of running `/bin/sh` directly.

### Permission denied errors

**Solution**: Boxify requires root privileges. Run with `sudo`.

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
- Port/socket already in use
- Missing directories
- Insufficient permissions

### Container filesystem is empty

**Solution**: Ensure Alpine rootfs is properly extracted to `/var/lib/boxify/boxify-rootfs/`

## Cleanup

### Remove a container

Containers are currently ephemeral and clean up on exit via deferred functions in `boxify-init`.

### Stop the daemon

```bash
sudo systemctl stop boxifyd
```

### Clean up all container data

```bash
sudo rm -rf /var/lib/boxify/boxify-container/*
```

## Project Structure

```
boxify/
├── cmd/
│   ├── boxify/          # Client CLI
│   ├── boxify-init/     # Container init process
│   └── boxifyd/         # Daemon
├── pkg/
│   ├── cgroup/          # Cgroups v2 management
│   ├── container/       # Container/overlay filesystem
│   ├── daemon/          # Daemon handlers and types
│   └── network/         # Networking (bridge, veth, IP management)
├── config/              # Configuration structures
└── boxify.yaml          # Container configuration (user-created)
```

## Limitations

- Single container per `boxify run` command
- No image management (uses Alpine rootfs directly)
- No container persistence between runs
- No volume mounts (yet)
- No port forwarding configuration
- Limited to Linux systems with cgroups v2

## Contributing

This is an educational project demonstrating container runtime fundamentals. Feel free to experiment and extend!

## License

[Add your license here]

## References

- [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- [Control Groups v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)
- [Overlay Filesystem](https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html)
