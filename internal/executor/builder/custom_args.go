package builder

import "strings"

func (builder *CommandBuilder) addCustomArgs() {
	if builder.Options.CustomArgs != "" {
		arguments := strings.Fields(builder.Options.CustomArgs)
		builder.Arguments = append(builder.Arguments, arguments...)
	}
}
