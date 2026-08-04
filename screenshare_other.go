//go:build !darwin && !windows

package main

// No reliable cross-desktop way to detect screen sharing here; wiggle anyway.
func screenBeingShared() bool {
	return false
}
