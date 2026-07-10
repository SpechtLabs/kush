package main

import (
	"context"
	"fmt"
	"os"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/cmd"
)

func main() {
	root := cmd.NewRootCmd()
	cmd.AddSubcommands(root)
	if err := root.ExecuteContext(context.Background()); err != nil {
		// humane errors render their advice; plain errors print bare.
		var herr humane.Error
		if errorsAs(err, &herr) {
			fmt.Fprintln(os.Stderr, herr.Display())
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}

// errorsAs is a tiny local alias so the humane type assertion reads clearly.
func errorsAs(err error, target *humane.Error) bool {
	h, ok := err.(humane.Error)
	if ok {
		*target = h
	}
	return ok
}
