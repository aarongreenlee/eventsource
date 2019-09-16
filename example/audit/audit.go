package audit

import (
	"errors"
	"time"
)

// AuditCreation stores creation data within events.
type Create struct {
	Event       string
	Created     time.Time
	CreatedBy   string `json:"createdBy"`
	CreatedByID string `json:"createdByID"`
}

// Validate confirms Create has been populated and contains no zero values.
func (c Create) Validate() error {
	var t time.Time
	switch {
	case c.Created == t, c.CreatedBy == "", c.CreatedByID == "":
		return errors.New("programming error: audit.Create has zero values")
	case c.Event == "":
		return errors.New("programming error: audit.Create has no event which is typically set by the command's apply implementation")
	}

	return nil
}
