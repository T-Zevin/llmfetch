package main

import (
	"fmt"
	"os"

	"github.com/xzw/llmfetch/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "llmfetch:", err)
		os.Exit(1)
	}
}
