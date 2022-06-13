package redis

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"github.com/NubeIO/rubix-automater/pkg/helpers/uuid"
	"github.com/go-redis/redis/v8"
	"sort"
)

const automaterTransaction = "automater-transaction"

// CreateTransaction adds a new trans result to the storage.
func (inst *Redis) CreateTransaction(job *model.Job) (*model.Transaction, error) {
	id, _ := uuid.New().Make("tra")
	key := inst.getRedisKeyForTransaction(id)
	now := ttime.New().Now()
	isPipeLine := false
	runAtUUID := "false"
	if job.PipelineID != "" {
		isPipeLine = true
		getPipeline, err := inst.GetPipeline(job.PipelineID)
		if err != nil {
			return nil, err
		}
		runAtUUID = getPipeline.RunAtUUID
	}
	trans := &model.Transaction{
		UUID:          id,
		PipelineID:    job.PipelineID,
		JobID:         job.UUID,
		TaskType:      job.TaskName,
		IsPipeLine:    isPipeLine,
		RunAtUUID:     runAtUUID,
		Status:        job.Status,
		FailureReason: job.FailureReason,
		StartedAt:     job.StartedAt,
		CreatedAt:     &now,
		CompletedAt:   job.CompletedAt,
		Duration:      job.Duration,
	}

	value, err := json.Marshal(trans)
	if err != nil {
		return nil, err
	}
	err = inst.Set(ctx, key, value, 0).Err()
	if err != nil {
		return nil, err
	}
	tran, err := inst.GetTransaction(id)
	if err != nil {
		return nil, err
	}

	pubTrans := model.PublishTransaction{
		UUID:          tran.UUID,
		JobID:         job.UUID,
		IsPipeLine:    isPipeLine,
		PipelineID:    job.PipelineID,
		TaskType:      tran.TaskType,
		Status:        tran.Status.String(),
		RunAtUUID:     tran.RunAtUUID,
		FailureReason: tran.FailureReason,
		CreatedAt:     tran.CreatedAt,
		StartedAt:     tran.StartedAt,
		CompletedAt:   tran.CompletedAt,
		Duration:      tran.Duration,
	}
	err = inst.Pub(automaterTransaction, pubTrans)
	if err != nil {
		return nil, err
	}

	return tran, nil
}

// DeleteTransaction deletes a trans from the storage.
func (inst *Redis) DeleteTransaction(uuid string) error {
	key := inst.getRedisKeyForTransaction(uuid)
	_, err := inst.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetTransactions fetches all trans from the storage, optionally filters the jobs by status.
func (inst *Redis) GetTransactions(status model.JobStatus) ([]*model.Transaction, error) {
	var keys []string
	key := inst.GetRedisPrefixedKey(fmt.Sprintf("%s:*", transaction))
	iter := inst.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	var jobs []*model.Transaction
	for _, key := range keys {
		value, err := inst.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		j := &model.Transaction{}
		if err := json.Unmarshal(value, j); err != nil {
			return nil, err
		}
		if status == model.Undefined || j.Status == status {
			jobs = append(jobs, j)
		}
	}
	// ORDER BY created_at ASC
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(*jobs[j].CreatedAt)
	})
	return jobs, nil
}

// GetTransactionsByJob fetches all trans from the storage by job
func (inst *Redis) GetTransactionsByJob(jobId string) ([]*model.Transaction, error) {
	var jobs []*model.Transaction
	transactions, err := inst.GetTransactions(model.Undefined)
	if err != nil {
		return nil, err
	}
	for _, t := range transactions {
		if t.JobID == jobId {
			jobs = append(jobs, t)
		}
	}
	return jobs, nil
}

// GetTransaction fetches a trans from the storage.
func (inst *Redis) GetTransaction(uuid string) (*model.Transaction, error) {
	key := inst.getRedisKeyForTransaction(uuid)
	val, err := inst.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, &apperrors.NotFoundErr{UUID: uuid, ResourceName: "transactionctl"}
		}
		return nil, err
	}
	var j *model.Transaction
	err = json.Unmarshal(val, &j)
	if err != nil {
		return nil, err
	}
	return j, nil
}
