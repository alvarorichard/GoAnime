// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:build windows

package ytdlp

import (
	"debug/pe"
	"os"
	"os/exec"
	"syscall"
)

// applySyscall applies any OS-specific syscall attributes to the command.
func applySyscall(cmd *exec.Cmd, separateProcessGroup bool) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW.
		HideWindow:    true,
	}
	if separateProcessGroup {
		cmd.SysProcAttr.CreationFlags |= syscall.CREATE_NEW_PROCESS_GROUP
	}
}

func isExecutable(path string, _ os.FileInfo) bool {
	// Try to parse as PE (Portable Executable) format.
	f, err := pe.Open(path)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
