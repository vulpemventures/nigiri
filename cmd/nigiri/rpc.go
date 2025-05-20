package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
)

var rpc = cli.Command{
	Name:   "rpc",
	Usage:  "invoke bitcoin-cli or elements-cli",
	Action: rpcAction,
	Flags: []cli.Flag{
		&liquidFlag,
		&cli.StringFlag{
			Name:  "rpcwallet",
			Usage: "rpcwallet to be used for node JSONRPC commands",
			Value: "",
		},
		&cli.IntFlag{
			Name:  "generate",
			Usage: "generate block",
			Value: 0,
		},
		&cli.BoolFlag{
			Name:  "named",
			Usage: "use named arguments",
			Value: false,
		},
	},
}

func rpcAction(ctx *cli.Context) error {
	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	isLiquid := ctx.Bool("liquid")
	rpcWallet := ctx.String("rpcwallet")
	generate := ctx.Int("generate")
	named := ctx.Bool("named")

	rpcArgs := []string{"exec", "bitcoin", "bitcoin-cli", "-rpcuser=admin1", "-rpcpassword=123", "-rpcport=18443", "-rpcwallet=" + rpcWallet}
	if isLiquid {
		rpcArgs = []string{"exec", "liquid", "elements-cli", "-datadir=config", "-rpcwallet=" + rpcWallet}
	}
	if generate > 0 {
		rpcArgs = append(rpcArgs, "-generate", fmt.Sprint(generate))
	}
	if named {
		rpcArgs = append(rpcArgs, "-named")
	}
	cmdArgs := append(rpcArgs, ctx.Args().Slice()...)
	bashCmd := exec.Command("docker", cmdArgs...)
	// Create a pipe for the output of the "docker exec" command
	r, w := io.Pipe()
	bashCmd.Stdout = w
	bashCmd.Stderr = os.Stderr

	// Start a goroutine to run the "docker exec" command
	go func() {
		if err := bashCmd.Run(); err != nil {
			w.CloseWithError(err)
		} else {
			w.Close()
		}
	}()

	// Read the output of the "docker exec" command from the pipe
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.Bytes()

	// Use the json.Unmarshal function to parse the output of the
	// "docker exec" command and check if it is a valid JSON object
	var v interface{}
	if err := json.Unmarshal(output, &v); err == nil {
		// Use the json.Marshal function to convert the parsed JSON object
		// to a byte slice
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}

		// Use the bytes.Buffer type to create a buffer that we can use
		// to write the indented JSON string to
		var buf bytes.Buffer

		// Use the json.Indent function to add indentation to the JSON byte slice
		// in the same way as if you were using the "jq" command
		if err := json.Indent(&buf, jsonBytes, "", "    "); err != nil {
			return err
		}

		// Split the indented JSON string into individual lines
		lines := strings.Split(buf.String(), "\n")

		// Loop through each line in the indented JSON string
		for _, line := range lines {
			// Check if the line starts with a "{" or a "["
			if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "[") {
				// If the line starts with a "{" or a "[", it is the start of a JSON object
				// or array, so print it without any color
				fmt.Println(line)
			} else if strings.Contains(line, ":") {
				// If the line contains a ":", it is a key-value pair, so split the line
				// into the key and value parts and add the desired colors to each part
				parts := strings.SplitN(line, ":", 2)
				key := parts[0]
				value := parts[1]
				// Use the AnsiColorCode function from the "github.com/logrusorgru/aurora"
				// package to create ANSI escape codes for the "key" and "value" colors
				keyColor := aurora.BrightBlue(key)
				valueColor := aurora.BrightCyan(value)

				fmt.Printf("%s: %s\n", keyColor.String(), valueColor.String())
			} else {
				// If the line does not start with a "{" or a "[" and does not contain a ":",
				// it is not a JSON object or array and does not contain a key-value pair, so
				// print it without any color
				fmt.Println(line)
			}
		}
	} else {
		fmt.Println(string(output))
	}
	return nil
}
