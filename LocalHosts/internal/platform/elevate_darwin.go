//go:build darwin

package platform

import (
	"errors"
)

func ApplyHostsWithAdmin(hostsPath string, content []byte) error {
	if err := applyHostsViaSession(hostsPath, content); err == nil {
		return nil
	} else if err != nil && !errors.Is(err, ErrNoElevatedSession) {
		return err
	}

	if err := RequestAdminAccess(); err != nil {
		return err
	}
	return applyHostsViaSession(hostsPath, content)
}
