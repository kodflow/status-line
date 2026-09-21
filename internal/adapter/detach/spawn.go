// Package detach re-executes this binary in the background to refresh a cache
// out of band.
package detach

import (
	"os"
	"os/exec"
)

// Spawn re-executes this binary with a flag, detached from the current process.
// A goroutine would die with the process, which exits as soon as the status
// line is printed, so the refresh has to outlive it.
//
// Params:
//   - flag: command-line flag the child runs with
//   - guardEnv: variable marking the child, so a refresh never spawns another
func Spawn(flag, guardEnv string) {
	// A refresh process must never spawn another one
	if os.Getenv(guardEnv) != "" {
		return
	}
	self, err := os.Executable()
	// Without a resolvable path there is nothing to re-execute
	if err != nil {
		return
	}
	cmd := exec.Command(self, flag)
	cmd.Env = append(os.Environ(), guardEnv+"=1")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	// Detach so the parent can exit immediately
	cmd.SysProcAttr = attr()
	// A refresh that cannot start simply leaves the cache stale
	if err := cmd.Start(); err != nil {
		return
	}
	// Release the child so it is not left as a zombie
	_ = cmd.Process.Release()
}
