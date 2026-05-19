package builder

import (
	"fmt"
	"light-launcher/internal/config"
	"path/filepath"
)

func (builder *CommandBuilder) buildBaseEnvironment() {
	builder.Environment = append(builder.Environment,
		fmt.Sprintf("WINEPREFIX=%s", config.ExpandPath(builder.Options.PrefixPath)),
	)

	if builder.Options.ProtonPath != "" {
		protonPattern := filepath.Base(builder.Options.ProtonPath)
		builder.Environment = append(builder.Environment,
			fmt.Sprintf("UMU_PROTON_PATTERN=%s", protonPattern),
			fmt.Sprintf("PROTONPATH=%s", config.ExpandPath(builder.Options.ProtonPath)),
		)
	}
}
