// Package version provides application version information.
// The version can be set at build time using ldflags:
//
//	go build -ldflags "-X github.com/ramonehamilton/MTGA-Companion/internal/version.Version=v1.2.3"
package version

// Version is the application version. It defaults to "dev" and can be
// overridden at build time using ldflags.
var Version = "dev"

// GetVersion returns the current application version.
func GetVersion() string {
	return Version
}
