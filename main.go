package main

import (
	"flag"
	"fmt"
	"os"
)

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

// @TODO add taring functionality that take the folder name as the tar name if the second flag isnt provided
// @TODO add a way for user to exclude the files they dont want to tar
