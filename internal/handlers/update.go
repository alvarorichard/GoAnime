package handlers

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/updater"
	"github.com/alvarorichard/Goanime/internal/util"
)

var (
	handleUpdateInitLogger       = util.InitLogger
	handleUpdateInfo             = util.Info
	handleCheckAndPromptUpdateFn = updater.CheckAndPromptUpdate
)

// HandleUpdateRequest processes update requests
func HandleUpdateRequest() error {
	// Initialize logger for update process
	handleUpdateInitLogger()
	handleUpdateInfo("Checking for updates...")
	if updateErr := handleCheckAndPromptUpdateFn(); updateErr != nil {
		return fmt.Errorf("update failed: %w", updateErr)
	}
	return nil
}
