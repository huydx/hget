package main

import (
	"io/ioutil"
	"path/filepath"
	"os"
	"fmt"
	"strings"
)

func TaskPrint() error {
	downloading, err := ioutil.ReadDir(filepath.Join(os.Getenv("HOME"), dataFolder))
	if err != nil {
		return err
	}

	folders := make([]string, 0)
	for _, d := range downloading {
		if d.IsDir() {
			folders = append(folders, d.Name())
		}
	}

	folderString := strings.Join(folders, "\n")
	Printf("Currently on going download: \n")
	fmt.Println(folderString)

	return nil
}

func Resume(task string) error {
	state, err := Read(task)
	if err != nil {
		return err
	}
	fmt.Println(state)
	return nil
}
