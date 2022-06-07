package automater

import (
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/timeconversion"
	"github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"github.com/jmattheis/go-timemath"
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

func PipelineRunAt(scheduleAt string, p *model.PipelineOptions, index int) (time.Time, error) {
	now := ttime.New().Now()
	var addDelay bool
	if p != nil && index > 0 {
		if p.DelayBetweenTask <= 0 {
			p.DelayBetweenTask = 1
		}
		if p.DelayBetweenTask > 0 {
			addDelay = true
		}
	}
	if scheduleAt == "" {
		scheduleAt = "1s"
	}
	if scheduleAt != "" {
		timeChecked, err := time.Parse(time.RFC3339Nano, scheduleAt)
		if err != nil {
			nextRunTime, err := timeconversion.AdjustTime(ttime.New().Now(), scheduleAt) // user can just pass in 15 sec on the HTTP POST RunAt
			if addDelay {
				nextRunTime = timemath.Second.Add(nextRunTime, p.DelayBetweenTask*index)
			}
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
