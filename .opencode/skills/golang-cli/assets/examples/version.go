// Path: internal/version/version.go
//
// Build with:
//
//	go build -ldflags "-X github.com/you/myapp/internal/version.Version=v1.2.3 \
//	  -X github.com/you/myapp/internal/version.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/you/myapp/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
//
// goreleaser sets the same flags from the git tag.
package version

import "runtime/debug"

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit,omitempty"`
	Date    string `json:"date,omitempty"`
	Go      string `json:"go"`
}

// Get reports build metadata. `go install module@version` builds carry no
// ldflags, so it falls back to the module version and VCS stamps Go embeds.
func Get() Info {
	info := Info{Version: Version, Commit: Commit, Date: Date}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}
	info.Go = bi.GoVersion
	if info.Version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		info.Version = bi.Main.Version
	}
	for _, s := range bi.Settings {
		switch {
		case s.Key == "vcs.revision" && info.Commit == "":
			info.Commit = s.Value
		case s.Key == "vcs.time" && info.Date == "":
			info.Date = s.Value
		}
	}
	return info
}
