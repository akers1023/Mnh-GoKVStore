# 🪟 Windows Quick Start Guide

Since you're on Windows, this container engine **WON'T RUN directly** on Windows. Here's how to run it:

## ✅ Recommended: Use WSL2

### 1. Install WSL2
Open PowerShell as Administrator:
```powershell
wsl --install
```

Restart your computer when prompted.

### 2. Open WSL2
Press `Win + R`, type `wsl`, press Enter.

### 3. Setup Project in WSL

```bash
# Navigate to your project
cd /mnt/d/Mnh-GoKVStore/mnh-go-container-engine

# Download Alpine Linux rootfs
mkdir -p ~/container-rootfs
cd ~/container-rootfs
wget http://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.0-x86_64.tar.gz
tar -xzf alpine-minirootfs-3.19.0-x86_64.tar.gz
rm alpine-minirootfs-3.19.0-x86_64.tar.gz
```

### 4. Update Config

Edit `cmd/engine/main.go`:
```go
const DefaultRootFS = "/home/YOUR_USERNAME/container-rootfs"
```

Get your username:
```bash
echo $HOME
```

### 5. Build & Run

```bash
cd /mnt/d/Mnh-GoKVStore/mnh-go-container-engine
go build -o container-engine cmd/engine/main.go
sudo ./container-engine run /bin/sh
```

Done! You're now running a container on Windows via WSL2.

## 📚 What This Project Does

Creates isolated Linux environments (containers) using:
- Linux namespaces (UTS, PID, Mount)
- Chroot for filesystem isolation
- Process management

## 🔄 Development Workflow

1. **Edit files in Windows** - Your IDE works normally
2. **Test in WSL2** - Run commands in WSL terminal
3. **Files sync automatically** - Changes appear in both

## 📖 More Details

- Full guide: `WSL2-SETUP.md`
- Main README: `README.md`
- Architecture: See `README.md` for how containers work

## 🆘 Need Help?

- Error: "operation not permitted" → Use `sudo`
- Error: "go: command not found" → Install Go in WSL
- Files not found? → Check path in `main.go`

## 🎯 Alternative: Use Docker Desktop

If you just want to use containers without building one:

1. Install Docker Desktop for Windows
2. Download from: https://www.docker.com/products/docker-desktop
3. Use Docker instead

But this project is for learning how Docker works internally!

