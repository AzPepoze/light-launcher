package builder

import (
	"fmt"
	"light-launcher/internal/types"
	"path/filepath"
	"strings"
)

func FormatCommandForDisplay(commandArguments []string, options types.LaunchOptions) string {
	var builder strings.Builder
	if options.Extras.Memory.Enabled && options.Extras.Memory.Value != "" {
		builder.WriteString(fmt.Sprintf("[MemMin:%s] ", options.Extras.Memory.Value))
	}
	builder.WriteString("WINEPREFIX=" + options.PrefixPath + " ")
	if options.ProtonPath != "" {
		builder.WriteString("UMU_PROTON_PATTERN=" + filepath.Base(options.ProtonPath) + " ")
	}
	if options.Extras.EnableMangoHud {
		builder.WriteString("MANGOHUD=1 ")
	}
	builder.WriteString(strings.Join(commandArguments, " "))
	return builder.String()
}
