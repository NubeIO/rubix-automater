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

// CreateTransaction adds a new trans result to the storage.
func (rs *Redis) CreateTransaction(job *model.Job) (*model.Transaction, error) {
	id, _ := uuid.New().Make("tra")
	key := rs.getRedisKeyForTransaction(id)
	now := ttime.New().Now()
	trans := &model.Transaction{
		UUID:          id,
		JobID:         job.UUID,
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
	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return nil, err
	}
	transaction, err := rs.GetTransaction(key)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// DeleteTransaction deletes a trans from the storage.
func (rs *Redis) DeleteTransaction(uuid string) error {
	key := rs.getRedisKeyForTransaction(uuid)
	_, err := rs.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetTransactions fetches all trans from the storage, optionally filters the jobs by status.
func (rs *Redis) GetTransactions(status model.JobStatus) ([]*model.Transaction, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey(fmt.Sprintf("%s:*", transaction))
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	var jobs []*model.Transaction
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		j := &model.Transaction{}
		if err := json.Unmarshal(value, j); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
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
func (rs *Redis) GetTransactionsByJob(jobId string) ([]*model.Transaction, error) {
	var jobs []*model.Transaction
	transactions, err := rs.GetTransactions(model.Undefined)
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
func (rs *Redis) GetTransaction(uuid string) (*model.Transaction, error) {
	key := rs.getRedisKeyForTransaction(uuid)
	val, err := rs.Get(ctx, key).Bytes()
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
