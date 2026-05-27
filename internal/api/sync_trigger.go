package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CCSwitchTrigger triggers cc-switch synchronization.
type CCSwitchTrigger interface {
	TriggerCCSwitch() error
}

// TriggerCCSwitchSync handles POST /api/v1/sync/ccswitch.
func TriggerCCSwitchSync(trigger CCSwitchTrigger) gin.HandlerFunc {
	return func(c *gin.Context) {
		go func() {
			if err := trigger.TriggerCCSwitch(); err != nil {
				log.Printf("cc-switch sync: %v", err)
			}
		}()
		c.JSON(http.StatusAccepted, gin.H{"status": "accepted", "source": "ccswitch"})
	}
}
