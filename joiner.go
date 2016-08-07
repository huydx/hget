package main

import (
	"os"
	"io"
)

func JoinFile(files []string, out string) error {
	outf, err := os.OpenFile(out, os.O_CREATE | os.O_WRONLY, 0600)
	defer outf.Close()
	if err != nil {
		return err
	}

	for _, f := range files {
		if err = copy(f, outf); err != nil {
			return err
		}
	}

	return nil
}

//this function split just to use defer
func copy(from string, to io.Writer) error {
	f, err := os.OpenFile(from, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	io.Copy(to, f)
	return nil
}

