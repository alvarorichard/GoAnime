//go:build windows

package player

import (
	"fmt"
	"os/exec"

	"github.com/alvarorichard/Goanime/internal/util"
)

func setProcessGroup(cmd *exec.Cmd) {
	// mensage debug
	if util.IsDebug {
		fmt.Println("Setting process group for command:", cmd.String())
	}

}
