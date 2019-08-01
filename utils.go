package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createTarball() error {
	flag.Parse()

	sourcedir := flag.Arg(1)
	if sourcedir == "" {
		fmt.Println("Please specify the source\nUsage: go-tarball source destinationfile.tar.gz")
		os.Exit(1)
	}

	absSourcePath, err := filepath.Abs(sourcedir) // getting absolute path to the directory to tar
	checkError(err)

	var absDestinationPath string

	destinationfile := flag.Arg(2)
	if destinationfile == "" {
		// take filename as destination name if no second flag
		absDestinationPath = absSourcePath + ".tar.gz"
	} else {
		destinationdir, err := filepath.Abs(destinationfile) // getting absolute path to the destination file
		absDestinationPath = destinationdir
		checkError(err)
	}

	// creating destination directory
	destinationDir, err := os.Create(absDestinationPath) // created file
	if err != nil {
		return err
	}

	var fileWriter io.WriteCloser = destinationDir
	fileWriter = gzip.NewWriter(destinationDir) // add gzip filter
	defer fileWriter.Close()
	tarWriter := tar.NewWriter(fileWriter)

	// going over the files
	err = filepath.Walk(absSourcePath, func(file string, fi os.FileInfo, err error) error {
		// for each file
		if err != nil {
			return err
		}
		// excluding
		if path.Ext(file) == ".tar" || path.Ext(file) == ".log" || strings.Contains(file, "node_module") || strings.Contains(file, ".git") {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(file, absSourcePath)
		if header.Name == "" {
			return nil
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !fi.IsDir() {
			// read file if its not a directory
			fmt.Println("Taring:", file)
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			if _, err := tarWriter.Write(data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
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
			if _, err := os.Stat(filename); err != nil {
				err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755
				checkError(err)
			}
		case tar.TypeReg:
			// handle normal file
			fmt.Println("Untarring :", filename)
			newFile, err := os.Create(filename)
			checkError(err)

			_, err = io.Copy(newFile, tarBallReader)
			checkError(err)

			err = os.Chmod(filename, os.FileMode(header.Mode))
			checkError(err)

			newFile.Close()
		default:
			fmt.Printf("Unable to untar type : %c in file %s", header.Typeflag, filename)
		}
	}

}
