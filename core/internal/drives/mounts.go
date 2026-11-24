package drives

import (
	"strings"
)

// mountEscapes decodes the octal escapes that /proc/self/mountinfo uses for special characters.
var mountEscapes = strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)

// mountPointFor returns the mount point that holds path, or "" when no mount matches.
// mountinfo is the text of /proc/self/mountinfo. The longest matching mount point wins.
func mountPointFor(mountinfo string, path string) string {
	best := ""
	for _, line := range strings.Split(mountinfo, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		mount := mountEscapes.Replace(fields[4])
		if containsPath(mount, path) && len(mount) > len(best) {
			best = mount
		}
	}
	return best
}

// containsPath reports whether mount is path itself or one of its parent folders.
func containsPath(mount string, path string) bool {
	if mount == "/" {
		return strings.HasPrefix(path, "/")
	}
	return path == mount || strings.HasPrefix(path, mount+"/")
}
