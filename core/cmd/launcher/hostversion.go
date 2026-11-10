package main

import "regexp"

// devHostVersion is what plugins see on builds that carry no release version.
const devHostVersion = "2.0.0"

var semverPrefix = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)`)

// hostVersion turns the build version into the plain semver that plugin manifests compare against.
// Tags such as "v2.1.0-3-gabc1234" keep their numbers; "dev" or a bare commit hash use the default.
func hostVersion(buildVersion string) string {
	match := semverPrefix.FindStringSubmatch(buildVersion)
	if match == nil {
		return devHostVersion
	}
	return match[1]
}
