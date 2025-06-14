/*
Copyright (c) 2025 Michael Lechner
This software is released under the MIT License.
See the LICENSE file for further details.
*/

package version

// Version information
const (
	Version   = "1.0.1b"
	BuildDate = "2025-06-03"
	Copyright = "© 2025 Michael Lechner"
)

// GetVersionInfo returns a formatted version string
func GetVersionInfo() string {
	return Version + " (" + BuildDate + ")"
}
