package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	inputPath = "testdata/input.txt"
	outPath   = "output.txt"
)

func TestCopy(t *testing.T) {
	testCases := []struct {
		limit        int64
		offset       int64
		outPath      string
		expectedPath string
	}{
		{
			limit:        0,
			offset:       0,
			outPath:      "output.txt",
			expectedPath: "testdata/out_offset0_limit0.txt",
		},
		{
			limit:        10,
			offset:       0,
			outPath:      "output.txt",
			expectedPath: "testdata/out_offset0_limit10.txt",
		},
		{
			limit:        10000,
			offset:       0,
			outPath:      "output.txt",
			expectedPath: "testdata/out_offset0_limit10000.txt",
		},
		{
			limit:        1000,
			offset:       100,
			outPath:      "output.txt",
			expectedPath: "testdata/out_offset100_limit1000.txt",
		},
		{
			limit:        1000,
			offset:       6000,
			outPath:      "output.txt",
			expectedPath: "testdata/out_offset6000_limit1000.txt",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.expectedPath, func(t *testing.T) {
			err := Copy(inputPath, outPath, tc.offset, tc.limit)
			if err != nil {
				return
			}
			outData, _ := os.ReadFile(tc.outPath)
			expectedData, _ := os.ReadFile(tc.expectedPath)
			assert.Equal(t, outData, expectedData)
		})
	}
}

func TestCopyInvalidLimit(t *testing.T) {
	err := Copy(inputPath, "out.txt", 0, -1)
	assert.ErrorAs(t, err, &ErrInvalidLimit)
}

func TestCopyInvalidOffset(t *testing.T) {
	err := Copy(inputPath, "out.txt", -1, 0)
	assert.ErrorAs(t, err, &ErrInvalidOffset)
}

func TestCopyOffsetMoreThanFile(t *testing.T) {
	err := Copy(inputPath, "out.txt", 100_000_000, 0)
	assert.ErrorAs(t, err, &ErrInvalidOffset)
}
