package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s /path/to/envdir command [args...]\n", os.Args[0])
		os.Exit(1)
	}

	dir := os.Args[1]
	cmd := os.Args[2:]
	envs, err := ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading envdir: %v\n", err)
		os.Exit(1)
	}

	code := RunCmd(cmd, envs)
	os.Exit(code)
}
