package version

// Version components
const (
	Maj = "1"
	Min = "1"
	Fix = "5"

	AppVer = 6
)

var (
	// Must be a string because scripts like dist.sh read this file.
	Version = "1.1.5"

	// GitCommit is the current HEAD set using ldflags.
	GitCommit string
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
