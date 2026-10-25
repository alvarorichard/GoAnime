package handlers

import (
	"fmt"

	"github.com/alvarorichard/Goanime/v2/internal/updater"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// HandleUpdateRequest processes update requests
func HandleUpdateRequest() error {
	// Initialize logger for update process
	util.InitLogger()
	util.Info("Checking for updates...")
	if updateErr := updater.CheckAndPromptUpdate(); updateErr != nil {
		return fmt.Errorf("update failed: %w", updateErr)
	}
	return nil
}
