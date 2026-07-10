package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	humane "github.com/sierrasoftworks/humane-errors-go"
	"github.com/spechtlabs/kush/internal/cmd"
)

func main() {
	root := cmd.NewRootCmd()
	cmd.AddSubcommands(root)
	if err := root.ExecuteContext(context.Background()); err != nil {
		var herr humane.Error
		if errors.As(err, &herr) {
			fmt.Fprintln(os.Stderr, herr.Display())
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}
