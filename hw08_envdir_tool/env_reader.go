package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	envs := make(Environment)
	for _, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			continue
		}

		entryName := entry.Name()
		if strings.Contains(entryName, "=") {
			return nil, fmt.Errorf("invalid file name %s", entryName)
		}

		file, err := os.Open(dir + "/" + entryName)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(file)
		var line string
		if scanner.Scan() {
			line = scanner.Text()
		} else {
			line = ""
		}
		line = strings.ReplaceAll(line, "\x00", "\n")
		line = strings.TrimRight(line, "\t ")
		envs[entryName] = EnvValue{
			Value:      line,
			NeedRemove: line == "",
		}
	}
	return envs, nil
}
