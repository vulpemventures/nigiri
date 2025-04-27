package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var forget = cli.Command{
	Name:   "forget",
	Usage:  "forget previously remembered flags for the start command",
	Action: forgetAction,
	// No flags needed for this command
}

func forgetAction(ctx *cli.Context) error {
	datadir := ctx.String("datadir")
	flagsFilePath := filepath.Join(datadir, savedFlagsFileName) // Use the same constant defined in start.go

	err := os.Remove(flagsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, it's already "forgotten"
			fmt.Println("No remembered flags found.")
			return nil
		}
		// For other errors (e.g., permission issues), return the error
		return fmt.Errorf("failed to remove saved flags file '%s': %w", flagsFilePath, err)
	}

	fmt.Printf("Successfully forgot flags stored in %s\n", flagsFilePath)
	return nil
}
