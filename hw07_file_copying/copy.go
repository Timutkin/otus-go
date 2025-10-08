package main

import (
	"errors"
	"io"
	"log"
	"os"

	//nolint:depguard
	"github.com/cheggaaa/pb/v3"
)

var (
	ErrOffsetExceedsFileSize       = errors.New("offset exceeds file size")
	ErrInvalidOffset               = errors.New("offset should be more or equal zero")
	ErrInvalidLimit                = errors.New("limit should be more or equal zero")
	batchSize                int64 = 1024
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if limit < 0 {
		return ErrInvalidLimit
	}
	if offset < 0 {
		return ErrInvalidOffset
	}
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer closeFile(fromFile)
	stat, err := fromFile.Stat()
	if err != nil {
		return err
	}
	if offset > stat.Size() {
		return ErrOffsetExceedsFileSize
	}

	toFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer closeFile(toFile)

	_, err = fromFile.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	if limit == 0 {
		limit = stat.Size()
	}

	bar := pb.Full.Start64(limit)
	progressBarWriter := bar.NewProxyWriter(toFile)
	var n int64
	for n < limit {
		if limit-n < batchSize {
			_, err = io.CopyN(progressBarWriter, fromFile, limit-n)
			n += limit - n
		} else {
			b, err := io.CopyN(progressBarWriter, fromFile, batchSize)
			n += b
			if err != nil {
				return err
			}
		}
	}
	if err != nil {
		return err
	}
	bar.Finish()
	return nil
}

func closeFile(f *os.File) {
	if f != nil {
		err := f.Close()
		if err != nil {
			log.Printf("failed to close file %s: %v", f.Name(), err)
		}
	}
}
