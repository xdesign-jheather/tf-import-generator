package main

import (
	"io"
	"os"
	"os/exec"
)

func format(r io.Reader) (string, error) {
	// Check terraform is available first.

	tf, err := exec.LookPath("terraform")

	if err != nil {
		return "", err
	}

	// Create a temp file on disk, so we can target with fmt.

	tmp, err := os.CreateTemp("", "tf-import-generator-format-*.tf")

	if err != nil {
		return "", err
	}

	// Ensure we close and delete the temp file once done.

	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	// Drain the reader into the temp file.

	if _, err := io.Copy(tmp, r); err != nil {
		return "", err
	}

	// Close it to ensure no more writes.

	_ = tmp.Close()

	// Run terraform fmt on the temp file.

	cmd := exec.Command(tf, "fmt", tmp.Name())

	if err = cmd.Run(); err != nil {
		return "", err
	}

	// Read the file back.

	bytes, err := os.ReadFile(tmp.Name())

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
