package service

import (
  "fmt"
  "mnhgo/mnh-go-container-engine/internal/container"
)

// CLI handles command-line interface for the container engine
type CLI struct {
  rootFsPath string
}

// NewCLI creates a new CLI instance
func NewCLI(rootFsPath string) *CLI {
  return &CLI{
    rootFsPath: rootFsPath,
  }
}

// Execute runs a command inside a container
func (c *CLI) Execute(command []string) error {
  if len(command) == 0 {
    return fmt.Errorf("no command specified")
  }
  
  return container.Run(command, c.rootFsPath)
}

// Help displays usage information
func Help() {
  fmt.Println(`
Mnh-Go-Container-Engine - A simple container runtime

USAGE:
  go run main.go run <command>

EXAMPLES:
  go run main.go run /bin/sh
  go run main.go run /bin/ls
  go run main.go run /bin/echo "Hello from container!"

LIMITATIONS:
  - This requires Linux kernel features (namespaces, cgroups)
  - On Windows, use WSL2 to run this container engine
  - On Mac, use a Linux VM or Docker Desktop

BUILDING:
  go build -o container-engine cmd/engine/main.go
  sudo ./container-engine run /bin/sh

NOTE:
  You need to provide a root filesystem. Download a minimal Linux rootfs
  from: https://alpinelinux.org/downloads/
  or extract from an existing Linux system.

  After downloading/extracting, update DefaultRootFS in main.go
`)
}
