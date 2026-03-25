// Package main is the entry point for the TelePromptr API server.
// It wires together all dependencies and starts the HTTP/gRPC server.
package main

import (
	"fmt"
	"os"

	"github.com/ralphlozano/telepromptr/apps/api/pkg/version"
)

func main() {
	fmt.Printf("TelePromptr API server starting... (version %s)\n", version.Version)
	os.Exit(0)
}
