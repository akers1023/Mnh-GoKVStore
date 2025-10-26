# Mnh-Go-Container-Engine

A minimal container runtime implementation in Go, similar to Docker's core concepts. This demonstrates Linux namespaces, mount points, and process isolation.

## ⚠️ Important: Windows Limitation

**This project requires Linux kernel features and WILL NOT RUN on Windows directly.**

### Why?
Container runtimes rely on Linux kernel features that don't exist on Windows:
- **Namespaces** (UTS, PID, Mount)
- **Cgroups** (resource limits)
- **Chroot** (filesystem isolation)

### Solutions for Windows Users:
1. **WSL2 (Windows Subsystem for Linux)** - Recommended
   ```bash
   wsl --install  # Install WSL2
   ```
2. **Docker Desktop** - Use Docker which provides this functionality
3. **Linux VM** - VirtualBox, VMware, or Hyper-V
4. **Cloud Linux Instance** - AWS, Azure, GCP

## What Does This Do?

This creates an isolated environment where:
- Processes have isolated process IDs
- Hostname can be changed independently
- Filesystem is isolated (chroot)
- New process namespace is created

It's similar to `docker run`, but much simpler and educational.

## Architecture

```
mnh-go-container-engine/
├── cmd/
│   └── engine/
│       └── main.go          # Parent/Child process coordinator
├── internal/
│   └── container/
│       ├── runtime.go       # Linux namespace isolation
│       └── fs.go            # Filesystem chroot setup
├── service/
│   └── cli.go              # CLI interface
└── README.md
```

## Prerequisites

### On Linux:
```bash
# Install Go
sudo apt install golang-go  # or your distro's package manager

# Make sure you have root privileges for namespace operations
```

### On Windows (via WSL2):
```bash
wsl --install
# Then follow Linux instructions above
```

## Setup Root Filesystem

You need a minimal Linux root filesystem. Options:

### Option 1: Alpine Linux (Recommended)
```bash
# Download Alpine Linux rootfs
cd /tmp
wget http://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.0-x86_64.tar.gz
mkdir -p ~/container-rootfs
tar -xzf alpine-minirootfs-3.19.0-x86_64.tar.gz -C ~/container-rootfs
```

### Option 2: Extract from Existing System
```bash
# Create rootfs directory
mkdir -p ~/container-rootfs/{bin,lib,usr,proc,dev,etc,tmp,home}

# Copy essential binaries
cp -r /bin/* ~/container-rootfs/bin/
cp -r /usr/bin/* ~/container-rootfs/usr/bin/

# Copy libraries
cp -r /lib/* ~/container-rootfs/lib/
```

### Option 3: Use BusyBox
```bash
mkdir -p ~/container-rootfs
cd ~/container-rootfs
wget https://busybox.net/downloads/binaries/1.35.0-x86_64-linux-musl/busybox
chmod +x busybox
./busybox --install .
```

## Configuration

Edit `cmd/engine/main.go` to set your root filesystem path:

```go
const DefaultRootFS = "/home/username/container-rootfs"  // Change this!
```

## Building

```bash
cd mnh-go-container-engine
go build -o container-engine cmd/engine/main.go
```

## Usage

### Run a Shell Inside Container

```bash
sudo ./container-engine run /bin/sh
```

Or if using `go run`:

```bash
sudo go run cmd/engine/main.go run /bin/sh
```

You should see:
- New hostname: `container-box`
- Isolated process IDs
- Isolated filesystem
- Can't see host's processes or files

### Run Other Commands

```bash
# List files
sudo ./container-engine run /bin/ls

# Echo something
sudo ./container-engine run /bin/echo "Hello from container!"

# Basic shell
sudo ./container-engine run /bin/sh
```

## How It Works

### 1. Parent Process (`run`)
- Creates a new process using `exec.Command` with namespace flags:
  - `CLONE_NEWUTS`: Isolated hostname
  - `CLONE_NEWPID`: Isolated process ID namespace
  - `CLONE_NEWNS`: Isolated mount namespace

### 2. Child Process (`child`)
- Runs inside the new namespaces
- Sets custom hostname: `container-box`
- Uses chroot to isolate filesystem
- Mounts `/proc` for proper PID namespace
- Executes the user's command

### 3. Process Flow

```
User runs: container-engine run /bin/sh
    |
    v
Parent (cmd/engine/main.go)
    |- Creates child with Linux namespaces
    |  (CLONE_NEWUTS | CLONE_NEWPID | CLONE_NEWNS)
    |
    v
Child Process
    |- Sets hostname to "container-box"
    |- Chroots to rootfs directory
    |- Mounts /proc
    |- Executes /bin/sh
```

## Code Walkthrough

### Namespaces
Linux namespaces provide isolation:
- **UTS namespace**: Isolated hostname (`syscall.CLONE_NEWUTS`)
- **PID namespace**: Isolated process IDs (`syscall.CLONE_NEWPID`)
- **Mount namespace**: Isolated filesystem (`syscall.CLONE_NEWNS`)

### Chroot
`chroot` changes the root filesystem directory, limiting the process's view of the filesystem to a subdirectory.

### Process Isolation
```go
// Parent creates child with namespaces
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
}
```

## Testing on Linux

### Interactive Test
```bash
sudo go run cmd/engine/main.go run /bin/sh

# Inside container:
hostname  # Should show: container-box
ps aux    # Should only show processes inside container
```

### Verify Isolation
```bash
# On host terminal:
ps aux | grep container  # You'll see host processes

# Inside container terminal:
ps aux                   # Container sees different PIDs
```

## Limitations

This is a **minimal educational implementation**. Missing features:

- ❌ Cgroups (resource limits)
- ❌ Network isolation
- ❌ Image management
- ❌ Layered filesystems
- ❌ Dockerfile support
- ❌ Port forwarding
- ❌ Volume mounts
- ❌ Multi-container orchestration

For production use, consider:
- Docker
- Podman
- containerd
- LXC

## Windows Users

### Using WSL2

1. Install WSL2:
   ```powershell
   wsl --install
   ```

2. Open WSL terminal and follow Linux instructions

3. In WSL:
   ```bash
   cd /mnt/d/Mnh-GoKVStore/mnh-go-container-engine
   go build -o container-engine cmd/engine/main.go
   sudo ./container-engine run /bin/sh
   ```

### Alternative: Docker Desktop
If you just want to use containers on Windows, Docker Desktop provides a full container runtime without needing this educational project.

## Why Build This?

Understanding how containers work helps you:
- Debug container issues
- Understand Docker's internals
- Learn Linux kernel features
- Build container-related tools
- Appreciate containerization complexity

## References

- Linux namespaces: https://man7.org/linux/man-pages/man7/namespaces.7.html
- Docker internals: https://docs.docker.com/get-started/overview/
- Rootless containers: https://rootlesscontaine.rs/

## License

MIT

