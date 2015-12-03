package memdb

import "errors"

var (
	ErrNoIDProvided  = errors.New("No ID provided")
	ErrIDNotExists   = errors.New("ID not exists")
	ErrIndexNotFound = errors.New("Secondary index not found")
)
