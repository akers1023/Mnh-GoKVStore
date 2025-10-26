# Running Mnh-Go-Container-Engine on Windows with WSL2

## Quick Start for Windows Users

### Step 1: Install WSL2 (if not already installed)

Open PowerShell as Administrator:

```powershell
wsl --install
```

This installs WSL2 with a default Linux distribution (usually Ubuntu).

### Step 2: Open WSL2 Terminal

Press `Win + R`, type `wsl`, and press Enter.

Or from PowerShell:
```powershell
wsl
```

### Step 3: Navigate to Project

```bash
cd /mnt/d/Mnh-GoKVStore/mnh-go-container-engine
```

(Adjust the path based on where your D: drive is mounted)

### Step 4: Setup Root Filesystem

```bash
# Create directory for rootfs
mkdir -p ~/container-rootfs
cd ~/container-rootfs

# Download Alpine Linux (minimal, ~2MB rootfs)
wget http://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.0-x86_64.tar.gz

# Extract
tar -xzf alpine-minirootfs-3.19.0-x86_64.tar.gz

# Clean up
rm alpine-minirootfs-3.19.0-x86_64.tar.gz

echo "Root filesystem ready at: $HOME/container-rootfs"
```

### Step 5: Update Configuration

Edit `cmd/engine/main.go`:

```go
const DefaultRootFS = "/home/your-username/container-rootfs"
```

Or get your username:
```bash
echo $HOME  # Use this path in DefaultRootFS
```

### Step 6: Build and Run

```bash
cd /mnt/d/Mnh-GoKVStore/mnh-go-container-engine

# Build
go build -o container-engine cmd/engine/main.go

# Run (requires sudo for namespace operations)
sudo ./container-engine run /bin/sh
```

### Step 7: Test Inside Container

Once inside the container:

```bash
# Check hostname
hostname  # Should show: container-box

# Check process isolation
ps aux    # Only shows container processes

# Basic commands
ls /
echo "Hello from container!"
```

## Troubleshooting

### Error: "operation not permitted"
- Make sure you're using `sudo`

### Error: "no such file or directory"
- Check that rootfs path exists and is correct
- Verify the directory structure: `ls ~/container-rootfs`

### Error: "permission denied"
- Some rootfs directories need proper permissions
- Run: `chmod -R 755 ~/container-rootfs`

### Build Error: "undefined: syscall.CLONE_NEWUTS"
- Make sure you're on Linux (WSL2 is fine)
- This won't compile on Windows

### Can't access /mnt/d
- WSL2 maps Windows drives to /mnt/
- Use `ls /mnt/` to see available drives
- Use the full path: `/mnt/d/Mnh-GoKVStore/...`

## Alternative: Use Docker Desktop

If you just want to learn containers without setup:

```powershell
# Install Docker Desktop for Windows
# Download from: https://www.docker.com/products/docker-desktop

# Then use Docker which provides full container runtime
docker run -it alpine sh
```

## Next Steps

1. Experiment with different commands
2. Try mounting directories
3. Learn Linux namespaces
4. Explore cgroups for resource limits

## Files Location

Your project files are at:
- Windows: `D:\Mnh-GoKVStore\mnh-go-container-engine`
- WSL2: `/mnt/d/Mnh-GoKVStore/mnh-go-container-engine`

You can edit files in Windows and they'll be accessible in WSL2.

