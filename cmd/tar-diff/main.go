package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/containers/tar-diff/pkg/common"
	tar_diff "github.com/containers/tar-diff/pkg/tar-diff"
)

var version = flag.Bool("version", false, "Show version")
var compressionLevel = flag.Int("compression-level", 3, "zstd compression level")
var maxBsdiffSize = flag.Int("max-bsdiff-size", 192, "Max file size in megabytes to consider using bsdiff, or 0 for no limit")

func main() {

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTION] old.tar.gz new.tar.gz result.tardiff\n", path.Base(os.Args[0]))
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *version {
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), common.VERSION)
		return
	}

	if flag.NArg() != 3 {
		flag.Usage()
		os.Exit(1)
	}

	oldFilename := flag.Arg(0)
	newFilename := flag.Arg(1)
	deltaFilename := flag.Arg(2)

	oldFile, err := os.Open(oldFilename)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	defer func() {
		if err := oldFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing %s: %s\n", oldFilename, err)
		}
	}()

	newFile, err := os.Open(newFilename)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer func() {
		if err := newFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing %s: %s\n", newFilename, err)
		}
	}()

	deltaFile, err := os.Create(deltaFilename)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	options := tar_diff.NewOptions()
	options.SetCompressionLevel(*compressionLevel)
	options.SetMaxBsdiffFileSize(int64(*maxBsdiffSize) * 1024 * 1024)

	err = tar_diff.Diff(oldFile, newFile, deltaFile, options)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	err = deltaFile.Close()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

}
