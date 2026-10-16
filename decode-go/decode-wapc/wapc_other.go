//go:build !wasip1

package main

import "errors"

// hostCall needs a waPC host, which only a wasip1 build has. This stub lets
// the package build and vet on the host.
func hostCall(_, _, _ string, _ []byte) ([]byte, error) {
	return nil, errors.New("waPC host calls need GOOS=wasip1")
}
