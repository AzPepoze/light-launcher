package systemsvc

import (
	"context"
	"errors"
)

var errUnsupported = errors.New("not supported on windows")

func (s *Service) lsfgInstalled() bool {
	return false
}

func (s *Service) shaderCacheBytes(_ context.Context) int64 {
	return 0
}

func (s *Service) clearShaderCache(_ context.Context) error {
	return errUnsupported
}

func (s *Service) dropCaches(_ context.Context) error {
	return errUnsupported
}

func (s *Service) clearSwap(_ context.Context) error {
	return errUnsupported
}

func (s *Service) cleanupProcesses(_ context.Context) error {
	return errUnsupported
}
