package main

import "os"

const (
	devtoolsEnv   = "LL_DEVTOOLS"
	disableGPUEnv = "LL_DISABLE_GPU"
	renderLogName = "render"
)

func developerTools() bool {
	return os.Getenv(devtoolsEnv) == "1"
}

func gpuDisabled() bool {
	return os.Getenv(disableGPUEnv) == "1"
}

func logRendering() {
	if gpuDisabled() {
		log.Warn(renderLogName, "GPU is off because %s=1", disableGPUEnv)
	}
}
