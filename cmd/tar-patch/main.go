package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/containers/tar-diff/pkg/common"
	tar_patch "github.com/containers/tar-diff/pkg/tar-patch"
)

var version = flag.Bool("version", false, "Show version")

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTION] file.tardiff /path/to/content destination.tar\n", path.Base(os.Args[0]))
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

	deltaFilename := flag.Arg(0)
	extractedDir := flag.Arg(1)
	patchedFilename := flag.Arg(2)

	dataSource := tar_patch.NewFilesystemDataSource(extractedDir)
	defer func() {
		if err := dataSource.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing %s: %s\n", extractedDir, err)
		}
	}()

	deltaFile, err := os.Open(deltaFilename)
	if err != nil {
		log.Fatalf("Unable to open %s: %s", deltaFilename, err)
	}
	defer func() {
		if err := deltaFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing %s: %s\n", deltaFilename, err)
		}
	}()

	var patchedFile *os.File

	if patchedFilename == "-" {
		patchedFile = os.Stdout
	} else {
		var err error
		patchedFile, err = os.Create(patchedFilename)
		if err != nil {
			log.Fatalf("Unable to create %s: %s", patchedFilename, err)
		}
		defer func() {
			if err := patchedFile.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing %s: %s\n", patchedFilename, err)
			}
		}()
	}

	err = tar_patch.Apply(deltaFile, dataSource, patchedFile)
	if err != nil {
		log.Fatalf("Error applying diff: %s", err)
	}
}
