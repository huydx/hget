package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

var dataFolder = ".hget/"
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
	//only working in unix with env HOME
	folder := filepath.Join(os.Getenv("HOME"), dataFolder, filepath.Base(s.Url))
	Printf("Saving current download data in %s\n", folder)
	if _, err := os.Stat(folder); err != nil {
		if err = os.MkdirAll(folder, 0700); err != nil {
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

func Read(task string) (*State, error) {
	file := filepath.Join(os.Getenv("HOME"), dataFolder, task, stateFileName)
	Printf("Getting data from %s\n", file)
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	s := new(State)
	err = json.Unmarshal(bytes, s)
	return s, err
}
