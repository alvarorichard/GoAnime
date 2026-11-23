package fs

import "fmt"

// ErrFileNotExist represents an error for when a file does not exist.
type ErrFileNotExist struct{ Msg string }

// Error returns the error message for ErrFileNotExist.
func (e *ErrFileNotExist) Error() string { return fmt.Sprintf("no such file: %s", e.Msg) }

// ErrFileClosed represents an error for when a file is already closed.
type ErrFileClosed struct{ Msg string }

// Error returns the error message for ErrFileClosed.
func (e *ErrFileClosed) Error() string {
	return fmt.Sprintf("%s: file is already closed and unlocked", e.Msg)
}
