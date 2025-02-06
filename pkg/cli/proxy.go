package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

const rootCmdName = "atmos"

var errAtmosCantBeExecuted = errors.New(`the command "atmos" can't be executed via atmos-proxy to prevent the infinite loop`)

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	cmdName := filepath.Base(args[0])
	if runtime.GOOS == "windows" {
		if e := strings.TrimSuffix(cmdName, ".exe"); e != cmdName {
			cmdName = e
		} else if e := strings.TrimSuffix(cmdName, ".bat"); e != cmdName {
			cmdName = e
		}
	}
	if cmdName == rootCmdName {
		fmt.Fprintln(os.Stderr, "[ERROR] "+errAtmosCantBeExecuted.Error())
		return errAtmosCantBeExecuted
	}
	cmd := exec.CommandContext(ctx, rootCmdName, append([]string{"exec", "--", cmdName}, args[1:]...)...) //nolint:gosec
	cmd.Stdin = runner.Stdin
	cmd.Stdout = runner.Stdout
	cmd.Stderr = runner.Stderr

	setCancel(cmd)

	if err := cmd.Run(); err != nil {
		return ecerror.Wrap(err, cmd.ProcessState.ExitCode())
	}
	return nil
}

const waitDelay = 1000 * time.Hour

func setCancel(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = waitDelay
}

func absoluteAquaPath() (string, error) {
	rootCmdPath, err := exec.LookPath(rootCmdName)
	if err != nil {
		return "", fmt.Errorf("%s isn't found: %w", rootCmdName, err)
	}
	if filepath.IsAbs(rootCmdPath) {
		return rootCmdPath, nil
	}
	a, err := filepath.Abs(rootCmdPath)
	if err != nil {
		return "", fmt.Errorf(`convert relative path "%s" to absolute path: %w`, rootCmdPath, err)
	}
	return a, nil
}
