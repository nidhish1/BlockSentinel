package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	LastBlock uint64 `json:"last_block"`
}

func resolveStateFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path does not exist; if it looks like a dir (ends with slash), create and use state.json inside
			if len(path) > 0 && (path[len(path)-1] == '/' || path[len(path)-1] == os.PathSeparator) {
				dir := path
				if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
					return "", mkErr
				}
				return filepath.Join(dir, "state.json"), nil
			}
			// Treat as file path
			return path, nil
		}
		return "", err
	}
	if info.IsDir() {
		return filepath.Join(path, "state.json"), nil
	}
	return path, nil
}

func loadState(path string) (uint64, error) {
	resolved, err := resolveStateFile(path)
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var state State
	json.Unmarshal(data, &state)
	return state.LastBlock, nil
}

func saveState(path string, blockNum uint64) error {
	resolved, err := resolveStateFile(path)
	if err != nil {
		return err
	}
	// Ensure parent dir exists
	if dir := filepath.Dir(resolved); dir != "." && dir != "" {
		if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
			return mkErr
		}
	}
	state := State{LastBlock: blockNum}
	data, _ := json.Marshal(state)
	return os.WriteFile(resolved, data, 0644)
}
