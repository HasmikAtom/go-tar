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

	// creating destination directory
	destinationDir, err := os.Create(absDestinationPath) // created file
	if err != nil {
		return err
	}

	var fileWriter io.WriteCloser = destinationDir

	// compressing if second flag contains "gz"
	if strings.HasSuffix(destinationfile, ".gz") {
		fileWriter = gzip.NewWriter(destinationDir) // add gzip filter
		defer fileWriter.Close()
	}

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
			// read file if its a directory
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
		if err := createTarball(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	if flag.Arg(0) == "untar" {
		extractTarball()
	}
}

// @TODO add a server

// @TODO specify extraction directory: tarball.tar.gz needs to be extracted in tarball directory :: DONE
// now the tarball is being extracted in the directory the bin file is being ran from
// @TODO add absolute path to the creating directories :: DONE

// @TODO add a way for user to exclude the files they dont want to tar
