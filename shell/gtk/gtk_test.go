//go:build linux

package main

import "testing"

func TestRenderingEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		disable   string
		devtools  string
		wantGPU   bool
		wantTools bool
	}{
		{name: "defaults", wantGPU: true},
		{name: "gpu off", disable: "1", wantGPU: false},
		{name: "gpu on", disable: "0", wantGPU: true},
		{name: "devtools on", devtools: "1", wantGPU: true, wantTools: true},
		{name: "devtools off", devtools: "0", wantGPU: true, wantTools: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(disableGPUEnv, test.disable)
			t.Setenv(devtoolsEnv, test.devtools)
			if got := !gpuDisabled(); got != test.wantGPU {
				t.Errorf("gpu enabled = %v, want %v", got, test.wantGPU)
			}
			if got := developerTools(); got != test.wantTools {
				t.Errorf("devtools = %v, want %v", got, test.wantTools)
			}
		})
	}
}
