package builder

import (
	"fmt"
	"light-launcher/core/internal/types"
	"path/filepath"
	"strings"
)

func FormatCommandForDisplay(commandArguments []string, options types.LaunchOptions) string {
	var builder strings.Builder
	if options.Extras.Memory.Enabled && options.Extras.Memory.Value != "" {
		builder.WriteString(fmt.Sprintf("[MemMin:%s] ", options.Extras.Memory.Value))
	}
	builder.WriteString("WINEPREFIX=")
	builder.WriteString(options.PrefixPath)
	builder.WriteString(" ")
	if options.ProtonPath != "" {
		builder.WriteString("UMU_PROTON_PATTERN=")
		builder.WriteString(filepath.Base(options.ProtonPath))
		builder.WriteString(" ")
	}
	if options.Extras.EnableMangoHud {
		builder.WriteString("MANGOHUD=1 ")
	}
	builder.WriteString(strings.Join(commandArguments, " "))
	return builder.String()
}
