package model

import (
	"fmt"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"strconv"
)

// JobStatus holds a value for job status ranging from 1 to 5.
type JobStatus int

const (
	Undefined  JobStatus = iota // 0
	Pending                     // 1
	Scheduled                   // 2
	InProgress                  // 3
	Completed                   // 4
	Failed                      // 5

	UNDERFINED = "UNDERFINED"
	PENDING    = "PENDING"
	SCHEDULED  = "SCHEDULED"
	INPROGRESS = "IN_PROGRESS"
	COMPLETED  = "COMPLETED"
	FAILED     = "FAILED"
)

// String converts the type to a string.
func (js JobStatus) String() string {
	if js != 0 {
		return [...]string{PENDING, SCHEDULED, INPROGRESS, COMPLETED, FAILED}[js-1]
	}
	return UNDERFINED

}

// Index returns the integer representation of a JobStatus.
func (js JobStatus) Index() int {
	return int(js)
}

// MarshalJSON for JSON representation.
func (js *JobStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + js.String() + `"`), nil
}

// UnmarshalJSON for JSON representation.
func (js *JobStatus) UnmarshalJSON(data []byte) error {
	var err error
	jobStatuses := map[string]JobStatus{
		PENDING:    Pending,
		SCHEDULED:  Scheduled,
		INPROGRESS: InProgress,
		COMPLETED:  Completed,
		FAILED:     Failed,
	}

	unquotedJobStatus, err := strconv.Unquote(string(data))
	if err != nil {
		return &apperrors.ResourceValidationErr{Message: fmt.Sprintf("unquoting job status data returned error: %s", err)}
	}

	jobStatus, ok := jobStatuses[unquotedJobStatus]
	if !ok {
		return &apperrors.ResourceValidationErr{Message: fmt.Sprintf("invalid status: %s", string(data))}
	}

	*js = jobStatus
	return err
}

// Validate makes a sanity check on JobStatus.
func (js JobStatus) Validate() error {
	var err error
	validJobStatuses := map[JobStatus]int{
		Pending:    Pending.Index(),
		Scheduled:  Scheduled.Index(),
		InProgress: InProgress.Index(),
		Completed:  Completed.Index(),
		Failed:     Failed.Index(),
	}
	if _, ok := validJobStatuses[js]; !ok {
		err = fmt.Errorf("%d is not a valid job status, valid statuses: %v", js, validJobStatuses)
	}
	return err
}
