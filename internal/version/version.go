package version

import "fmt"

const VERSION_MAJOR = 1
const VERSION_MINOR = 1
const VERISON_MICRO = 5

var version *Version

type Version struct {
	Major int
	Minor int
	Micro int
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERISON_MICRO)
}

func GetVersion() *Version {
	return version
}

func init() {
	version = new(Version)
	version.Major = VERSION_MAJOR
	version.Minor = VERSION_MINOR
	version.Micro = VERISON_MICRO
}
