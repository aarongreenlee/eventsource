// Package idgen produces IDs suitable for our example.
package idgen

import (
	"strconv"
	"time"
)

// NewID produces a new ID suitable for our examples.
func NewID() string {
	return strconv.Itoa(int(time.Now().Unix()))
}
