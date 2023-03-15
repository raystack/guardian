package utils

import (
	"fmt"
	"github.com/goto/guardian/domain"
	"time"
)

// GetReadableDuration returns a human-readable duration string in integer days preferably, or the original string if it's either not a valid duration or a days value is not integer.
func GetReadableDuration(durationStr string) (string, error) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return durationStr, err
	}

	days := duration.Hours() / 24
	if days > 0 {
		if IsInteger(days) {
			// if the duration is in integral days, return it as integer
			return fmt.Sprintf("%dd", int(days)), nil
		}

		return durationStr, nil
	}

	return domain.PermanentDurationLabel, nil
}
