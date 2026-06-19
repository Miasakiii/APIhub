package api

import "apihub/internal/util"

// generateID is a convenience wrapper around util.GenerateID().
func generateID() string {
	return util.GenerateID()
}
