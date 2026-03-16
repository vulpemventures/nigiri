package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"maps"
	"os"
	"path/filepath"
	"strconv"
)

type State struct {
	directory    string
	filePath     string
	initialState map[string]string
}

func New(filePath string, initialState map[string]string) *State {
	dir := filepath.Dir(filePath)
	return &State{
		directory:    dir,
		filePath:     filePath,
		initialState: initialState,
	}
}

func (s *State) FilePath() string {
	return s.filePath
}

func (s *State) Get() (map[string]string, error) {
	file, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := s.Set(s.initialState); err != nil {
			return nil, err
		}
		return s.initialState, nil
	}

	data := map[string]string{}
	json.Unmarshal(file, &data)

	return data, nil
}

func (s *State) Set(data map[string]string) error {
	if _, err := os.Stat(s.directory); os.IsNotExist(err) {
		os.Mkdir(s.directory, os.ModeDir|0755)
	}

	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	currentData, err := s.Get()
	if err != nil {
		return err
	}

	mergedData := merge(currentData, data)

	jsonString, err := json.Marshal(mergedData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(s.filePath, jsonString, 0755)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func (s *State) GetBool(key string) (bool, error) {
	stateData, err := s.Get()
	if err != nil {
		return false, err
	}

	if _, ok := stateData[key]; !ok {
		return false, errors.New("config: missing key " + key)
	}

	value, err := strconv.ParseBool(stateData[key])
	if err != nil {
		return false, err
	}

	return value, nil
}

func (s *State) GetString(key string) (string, error) {
	stateData, err := s.Get()
	if err != nil {
		return "", err
	}

	if _, ok := stateData[key]; !ok {
		return "", errors.New("config: missing key " + key)
	}

	return stateData[key], nil
}

func merge(mapsParam ...map[string]string) map[string]string {
	merge := make(map[string]string)
	for _, m := range mapsParam {
		maps.Copy(merge, m)
	}
	return merge
}
