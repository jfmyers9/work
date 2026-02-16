package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var ErrAborted = errors.New("editor exited with non-zero status")

// OpenEditor is the function used to edit content. Tests can replace this
// to avoid spawning a real editor.
var OpenEditor func(content, prefix, editorBin string) (string, error) = EditTempFile

// EditTempFile opens editorBin on a temp file seeded with content.
// Returns the edited content. The temp file is removed after reading.
func EditTempFile(content, prefix, editorBin string) (string, error) {
	f, err := os.CreateTemp("", prefix+"-*.md")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := f.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing temp file: %w", err)
	}

	path, err := exec.LookPath(editorBin)
	if err != nil {
		return "", fmt.Errorf("editor %q not found in PATH; set $EDITOR to your preferred editor", editorBin)
	}

	cmd := exec.Command(path, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", ErrAborted
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("reading temp file: %w", err)
	}
	return string(data), nil
}
