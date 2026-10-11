package main

import (
	"fmt"
	"os"

	"github.com/neoduck0/ivo/internal/config"
	"github.com/neoduck0/ivo/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := tui.Run(cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
