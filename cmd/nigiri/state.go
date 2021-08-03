package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

func getState() (map[string]string, error) {
	file, err := ioutil.ReadFile(statePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := setState(initialState); err != nil {
			return nil, err
		}
		return initialState, nil
	}

	data := map[string]string{}
	json.Unmarshal(file, &data)

	return data, nil
}

func setState(data map[string]string) error {
	if _, err := os.Stat(defaultDataDir); os.IsNotExist(err) {
		os.Mkdir(defaultDataDir, os.ModeDir|0755)
	}

	file, err := os.OpenFile(statePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	currentData, err := getState()
	if err != nil {
		return err
	}

	mergedData := merge(currentData, data)

	jsonString, err := json.Marshal(mergedData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(statePath, jsonString, 0755)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func merge(maps ...map[string]string) map[string]string {
	merge := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			merge[k] = v
		}
	}
	return merge
}

func getBoolFromState(key string) (bool, error) {
	stateConfig, err := getState()
	if err != nil {
		return false, err
	}

	if _, ok := stateConfig[key]; !ok {
		return false, errors.New("config: missing key " + key)
	}

	value, err := strconv.ParseBool(stateConfig[key])
	if err != nil {
		return false, err
	}

	return value, nil
}
