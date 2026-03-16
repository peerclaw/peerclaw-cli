//go:build !linux && !darwin

package cmd

import (
	"fmt"
	"runtime"
)

func newServiceManager() serviceManager {
	return &unsupportedServiceManager{}
}

type unsupportedServiceManager struct{}

func (u *unsupportedServiceManager) Install(_ serviceConfig) error {
	return fmt.Errorf("service management is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func (u *unsupportedServiceManager) Uninstall() error {
	return fmt.Errorf("service management is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func (u *unsupportedServiceManager) Status() (string, error) {
	return "", fmt.Errorf("service management is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
}

func (u *unsupportedServiceManager) Logs(_ int, _ bool) error {
	return fmt.Errorf("service management is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
}
