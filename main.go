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
	sourcedir := flag.Arg(1)

	if sourcedir == "" {
		fmt.Println("Please specify the source\nUsage: go-tarball source destinationfile.tar.gz")
		os.Exit(1)
	}

	destinationfile := flag.Arg(2)

	if destinationfile == "" {
		fmt.Println("Please specify the destination\nUsage: go-tarball source destinationfile.tar.gz")
		os.Exit(1)
	}

	absSourcePath, err := filepath.Abs(sourcedir) // getting absolute path to the directory to tar
	checkError(err)

	absDestinationPath, err := filepath.Abs(destinationfile) // getting absolute path to the destination file
	checkError(err)

	dir, err := os.Open(absSourcePath)
	checkError(err)

	defer dir.Close()

	files, err := dir.Readdir(0)
	checkError(err)

	tarfile, err := os.Create(absDestinationPath)
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

		header, err := tar.FileInfoHeader(fileInfo, "")
		checkError(err)

		err = tarfileWriter.WriteHeader(header)
		checkError(err)

		_, err = io.Copy(tarfileWriter, file)
		checkError(err)

	}
}

func extractTarball() {

	flag.Parse() // get the arguments from command line

	sourcefile := flag.Arg(1)

	if sourcefile == "" {
		fmt.Println("Usage : go-untar sourcefile.tar")
		os.Exit(1)
	}

	absSourcePath, err := filepath.Abs(sourcefile) // getting absolute path to the directory to tar
	checkError(err)

	file, err := os.Open(absSourcePath)
	checkError(err)

	defer file.Close()

	var fileReader io.ReadCloser = file

	// getting destination directory
	destinationdir := flag.Arg(2)

	var destEmpty bool
	var absDestinationPath string
	var extension string

	// setting the destination directory, if the flag is empty it gets the tar filename, of not it sets it to the flag value
	if destinationdir == "" {
		destEmpty = true
		var filename = file.Name()
		extension = filepath.Ext(filename)
		destinationdir := filename[0 : len(filename)-len(extension)]
		absDestinationPath = destinationdir

	} else {
		destEmpty = false
		destinationdir, err := filepath.Abs(destinationdir)
		absDestinationPath = destinationdir
		checkError(err)
	}

	// just in case we are reading a tar.gz file, add a filter to handle gzipped file
	if strings.HasSuffix(sourcefile, ".gz") {
		if fileReader, err = gzip.NewReader(file); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// trimming the gz extension if destination flag is empty
		if destEmpty {
			extension = filepath.Ext(absDestinationPath)
			absDestinationPath = absDestinationPath[0 : len(absDestinationPath)-len(extension)]
		}

		defer fileReader.Close()
	}

	//making the parent directory where the files will be untarred
	err = os.MkdirAll(absDestinationPath, os.ModePerm)
	checkError(err)

	tarBallReader := tar.NewReader(fileReader)

	parentDir, err := os.Open(absDestinationPath)
	defer parentDir.Close()

	// Extracting tarred files
	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			os.Exit(1)
		}

		// get the individual filename and extract to the current directory
		filename := filepath.Join(absDestinationPath, header.Name) // add absolute path
		switch header.Typeflag {
		case tar.TypeDir:
			// handle directory
			fmt.Println("Creating directory :", filename)

			err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755
			checkError(err)

		case tar.TypeReg:
			// handle normal file
			fmt.Println("Untarring :", filename)
			writer, err := os.Create(filename)

			checkError(err)

			io.Copy(writer, tarBallReader)

			err = os.Chmod(filename, os.FileMode(header.Mode))

			checkError(err)

			writer.Close()
		default:
			fmt.Printf("Unable to untar type : %c in file %s", header.Typeflag, filename)
		}
	}

}

func main() {

	flag.Parse()
	if flag.Arg(0) == "tar" {
		createTarball()
	}
	if flag.Arg(0) == "untar" {
		extractTarball()
	}
}

// @TODO add a server

// @TODO specify extraction directory: tarball.tar.gz needs to be extracted in tarball directory :: DONE
// now the tarball is being extracted in the directory the bin file is being ran from
// @TODO add absolute path to the creating directories :: DONE
