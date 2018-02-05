package handlers

import (
	"errors"
	"strings"

	"github.com/flavio/collage/types"
)

func translateName(mountsByReg types.MountPointsByRegistry, name string) (registry string, remoteName string, err error) {
	var source, target string

	for reg, mountPoints := range mountsByReg {
		for _, mp := range mountPoints {
			if strings.HasPrefix(name, mp.Target) {
				target = mp.Target
				source = mp.Source
				break
			}
		}
		if target != "" {
			registry = reg
			break
		}
	}
	if registry == "" || target == "" || source == "" {
		err = errors.New("Cannot find mount point")
		return
	}

	remoteName = strings.Replace(name, target, source, 1)

	return
}
