package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createTarball() {
	// parsing the flags
	flag.Parse()
	sourcedir := flag.Arg(0)

	if sourcedir == "" {
		fmt.Println("Please specify the source\nUsage: go-tarball source destinationfile.tar.gz")
		os.Exit(1)
	}

	destinationfile := flag.Arg(1)

	if destinationfile == "" {
		fmt.Println("Please specify the destination\nUsage: go-tarball source destinationfile.tar.gz")
		os.Exit(1)
	}

	dir, err := os.Open(sourcedir)

	checkError(err)

	defer dir.Close()

	files, err := dir.Readdir(0)

	checkError(err)

	tarfile, err := os.Create(destinationfile)

	checkError(err)
	defer tarfile.Close()
	var fileWriter io.WriteCloser = tarfile

	if strings.HasSuffix(destinationfile, ".gz") {
		fileWriter = gzip.NewWriter(tarfile) // add gzip filter
		defer fileWriter.Close()
	}

	tarfileWriter := tar.NewWriter(fileWriter)

	defer tarfileWriter.Close()

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}

		file, err := os.Open(dir.Name() + string(filepath.Separator) + fileInfo.Name())

		checkError(err)

		defer file.Close()

		header := new(tar.Header)
		header.Name = file.Name()
		header.Size = fileInfo.Size()
		header.Mode = int64(fileInfo.Mode())
		header.ModTime = fileInfo.ModTime()

		err = tarfileWriter.WriteHeader(header)

		checkError(err)

		_, err = io.Copy(tarfileWriter, file)

		checkError(err)

	}
}

func extractTarball() {

}

func main() {
	createTarball()
}
