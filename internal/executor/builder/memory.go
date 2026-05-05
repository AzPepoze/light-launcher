package builder

import (
	"fmt"
	"light-launcher/internal/system"
)

func (builder *CommandBuilder) applyMemoryProtection() {
	if builder.Options.Extras.Memory.Enabled && builder.Options.Extras.Memory.Value != "" && system.IsCommandAvailable("systemd-run") {
		wrappedArguments := []string{
			"systemd-run",
			"--user",
			"--scope",
			fmt.Sprintf("-pMemoryMin=%s", builder.Options.Extras.Memory.Value),
			"--",
		}
		builder.Arguments = append(wrappedArguments, builder.Arguments...)
	}
}
