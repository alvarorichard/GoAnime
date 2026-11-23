// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:build unix

package ytdlp

import (
	"os"
	"os/exec"
	"syscall"
)

// applySyscall applies any OS-specific syscall attributes to the command.
func applySyscall(cmd *exec.Cmd, separateProcessGroup bool) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: separateProcessGroup,
	}
}

func isExecutable(_ string, stat os.FileInfo) bool {
	// On Unix systems, check if executable bit is set (user, group, or others).
	return stat.Mode().Perm()&0o100 != 0 || stat.Mode().Perm()&0o010 != 0 || stat.Mode().Perm()&0o001 != 0
}
