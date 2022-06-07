package automater

import (
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/timeconversion"
	"github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"time"
)

func RunAt(runAt string) (time.Time, error) {
	now := ttime.New().Now()
	if runAt != "" {
		timeChecked, err := time.Parse(time.RFC3339Nano, runAt)
		if err != nil {
			nextRunTime, err := timeconversion.AdjustTime(ttime.New().Now(), runAt) // user can just pass in 15 sec on the HTTP POST RunAt
			if err != nil {
				return now, &apperrors.ParseTimeErr{Message: err.Error()}
			} else {
				return nextRunTime, nil
			}
		} else {
			return timeChecked, nil
		}
	} else {
		return now, nil
	}

}
