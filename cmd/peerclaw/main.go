package main

import (
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
