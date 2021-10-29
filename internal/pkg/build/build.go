package build

import (
	"runtime"
)

var (
	// Version current package version
	Version string

	// BuildDate current package build date
	BuildDate string
)

// GoVersion go version current package build with
var GoVersion = runtime.Version()
