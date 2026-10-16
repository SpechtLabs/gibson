// gibson renders a retro neon "SpechtLabs" intro sequence, locally or as an
// SSH app.
package main

import (
	"context"
	"os"

	"github.com/spechtlabs/gibson/cmd/gibson/command"
)

// version is set by GoReleaser.
var version = "dev"

func main() {
	os.Exit(command.Execute(context.Background(), version))
}
