package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func readState(r io.Reader) State {
	var state State

	if err := json.NewDecoder(r).Decode(&state); err != nil {
		log.Fatal(err)
	}

	return state
}

func uploadedStateFiles(r *http.Request, field string) (States, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	stateFiles := make(States, len(r.MultipartForm.File))

	for _, upload := range r.MultipartForm.File[field] {
		f, err := upload.Open()

		if err != nil {
			return nil, err
		}

		state := readState(f)

		_ = f.Close()

		if state.Version != 4 {
			continue
		}

		var resources Resources

		for _, resource := range state.Resources {
			if resource.Mode != "managed" {
				continue
			}

			resources = append(resources, resource)
		}

		if len(resources) == 0 {
			continue
		}

		state.Resources = resources

		stateFiles[shortenFilename(upload.Filename)] = state
	}

	if len(stateFiles) != len(r.MultipartForm.File[field]) {
		return nil, fmt.Errorf("duplicate or empty state files uploaded")
	}

	return stateFiles, nil
}

func shortenFilename(filename string) string {
	switch {
	case strings.HasSuffix(strings.ToLower(filename), ".json"):
		return shortenFilename(filename[:len(filename)-5])
	case strings.HasSuffix(strings.ToLower(filename), ".tfstate"):
		return shortenFilename(filename[:len(filename)-8])
	default:
		return filename
	}
}
