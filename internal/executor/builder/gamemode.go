package builder

import "light-launcher/internal/system"

func (builder *CommandBuilder) applyGameMode() {
	if builder.Options.Extras.EnableGamemode && system.IsCommandAvailable("gamemoderun") {
		builder.Arguments = append([]string{"gamemoderun"}, builder.Arguments...)
	}
}
