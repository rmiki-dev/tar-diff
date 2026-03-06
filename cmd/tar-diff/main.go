// Package main implements the tar-diff command line tool for creating binary diffs between tar archives.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/containers/tar-diff/pkg/protocol"
	tardiff "github.com/containers/tar-diff/pkg/tar-diff"
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
		fmt.Printf("%s %s\n", path.Base(os.Args[0]), protocol.VERSION)
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

	newFile, err := os.Open(newFilename)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	deltaFile, err := os.Create(deltaFilename)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	options := tardiff.NewOptions()
	options.SetCompressionLevel(*compressionLevel)
	options.SetMaxBsdiffFileSize(int64(*maxBsdiffSize) * 1024 * 1024)

	err = tardiff.Diff(oldFile, newFile, deltaFile, options)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

}
