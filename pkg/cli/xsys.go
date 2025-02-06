//go:build !windows
// +build !windows

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func (runner *Runner) RunXSysExec(args ...string) error {
	cmdName := filepath.Base(args[0])
	if cmdName == rootCmdName {
		return errAtmosCantBeExecuted
	}

	aquaPath, err := absoluteAquaPath()
	if err != nil {
		return fmt.Errorf("get %s's absolute path: %w", rootCmdName, err)
	}
	if err := unix.Exec(aquaPath, append([]string{rootCmdName, "exec", "--", cmdName}, args[1:]...), os.Environ()); err != nil {
		return fmt.Errorf("execute %s: %w", rootCmdName, err)
	}
	return nil
}
