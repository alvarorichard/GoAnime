// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//go:build !windows && !unix

package ytdlp

import (
	"os"
	"os/exec"
)

// applySyscall applies any OS-specific syscall attributes to the command.
func applySyscall(_ *exec.Cmd, _ bool) {
	// No-op by default.
}

func isExecutable(_ string, stat os.FileInfo) bool {
	return true // no-op.
}
