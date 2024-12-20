package main

import (
	"net/http"
)

func uploadedStateFiles(r *http.Request, field string) ([]State, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	var stateFiles []State

	for _, upload := range r.MultipartForm.File[field] {
		f, err := upload.Open()

		if err != nil {
			return nil, err
		}

		state := readState(f)

		_ = f.Close()

		stateFiles = append(stateFiles, state)
	}

	return stateFiles, nil
}
