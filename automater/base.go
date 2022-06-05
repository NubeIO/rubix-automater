package automater

import (
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"time"
)

func RunAt(runAt string) (time.Time, error) {
	now := ttime.New().Now()
	if runAt != "" {
		now, err := time.Parse(time.RFC3339Nano, runAt)
		if err != nil {
			return now, &apperrors.ParseTimeErr{Message: err.Error()}
		}
		return now, nil
	} else {
		return now, nil
	}
}
