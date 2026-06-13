// Package store manages database connectivity and queries for tessera-data.
package store

import "errors"

// ErrNotFound is returned when a queried record does not exist in the database.
var ErrNotFound = errors.New("not found")
