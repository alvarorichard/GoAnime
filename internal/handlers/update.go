package handlers

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/updater"
	"github.com/alvarorichard/Goanime/internal/util"
)

var checkAndPromptUpdate = updater.CheckAndPromptUpdate

// HandleUpdateRequest processes update requests
func HandleUpdateRequest() error {
	// Initialize logger for update process
	util.InitLogger()
	util.Info("Checking for updates...")
	if updateErr := checkAndPromptUpdate(); updateErr != nil {
		return fmt.Errorf("update failed: %w", updateErr)
	}
	return nil
}
