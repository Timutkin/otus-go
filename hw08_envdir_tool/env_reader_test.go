package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadDir(t *testing.T) {
	dir := filepath.Join("testdata", "env")

	t.Run("happy path", func(t *testing.T) {
		env, _ := ReadDir(dir)
		expectedEnvs := map[string]string{
			"BAR":   "bar",
			"EMPTY": "",
			"FOO": `   foo
with new line`,
			"HELLO": `"hello"`,
			"UNSET": "",
		}
		for k, eVal := range expectedEnvs {
			aVal, ok := env[k]
			assert.True(t, ok)
			assert.Equal(t, eVal, aVal.Value)
		}
	})

	t.Run("unset", func(t *testing.T) {
		err := os.Setenv("UNSET", "UNSET")
		if err != nil {
			return
		}
		env, _ := ReadDir(dir)
		val, ok := env["UNSET"]
		assert.True(t, ok)
		assert.Equal(t, "", val.Value)
		assert.True(t, val.NeedRemove)
	})
}
