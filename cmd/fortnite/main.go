package main

import (
	"os"

	"github.com/ygncode/fortnite-cli/internal/api"
	"github.com/ygncode/fortnite-cli/internal/commands"
)

// version is injected by ldflags at release time.
var version = "dev"

func main() {
	api.Version = version
	root := commands.NewRoot(version)
	if err := root.Execute(); err != nil {
		// Cobra prints its own usage errors before returning; we only need to
		// translate to the documented exit code.
		os.Exit(commands.ExitCodeFor(err))
	}
}
