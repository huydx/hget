package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

var dataFolder = "~/.hget/"
var stateFileName = "state.json"

type State struct {
	Url   string
	Parts []Part
}

type Part struct {
	Url       string
	Path      string
	RangeFrom int64
	RangeTo   int64
	Cursor    int64
}

func (s *State) Save() error {
	//make temp folder
	folder := filepath.Join(dataFolder, filepath.Base(s.Url))
	if _, err := os.Stat(folder); err != nil {
		if err = os.Mkdir(folder, 0600); err != nil {
			return err
		}
	}

	//move current downloading file to data folder
	for _, part := range s.Parts {
		os.Rename(part.Path, filepath.Join(folder, filepath.Base(part.Path)))
	}

	//save state file
	j, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(folder, stateFileName), j, 0644)
}

func (s *State) Resume() error {
	return nil
}
