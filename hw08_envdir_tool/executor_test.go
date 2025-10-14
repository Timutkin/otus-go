package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCmd_NoCommand(t *testing.T) {
	code := 0
	out := captureOutput(func() {
		code = RunCmd(nil, Environment{})
	})
	assert.Equal(t, 1, code)
	assert.Equal(t, "no command provided", out)
}

func TestRunCmd_AppliesEnvironmentAndRunsCommand(t *testing.T) {
	dir := filepath.Join("testdata", "env")
	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}
	env["ADDED"] = EnvValue{Value: "added"}

	_ = os.Setenv("HELLO", "original")
	_ = os.Setenv("BAR", "original")
	_ = os.Setenv("FOO", "original")
	_ = os.Setenv("UNSET", "should_be_removed_or_empty")
	_ = os.Setenv("EMPTY", "should_be_empty")
	_ = os.Unsetenv("ADDED")

	scriptPath := filepath.Join("testdata", "echo.sh")
	cmd := []string{"bash", scriptPath, "arg1", "arg2"}

	code := 0
	out := captureOutput(func() {
		code = RunCmd(cmd, env)
	})
	assert.Equal(t, 0, code)

	expected := "HELLO is (\"hello\")\n" +
		"BAR is (bar)\n" +
		"FOO is (   foo\nwith new line)\n" +
		"UNSET is ()\n" +
		"ADDED is (added)\n" +
		"EMPTY is ()\n" +
		"arguments are arg1 arg2\n"

	assert.Equal(t, expected, out)
}

func captureOutput(fn func()) string {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	defer func() {
		_ = w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	out := <-done
	return out
}
