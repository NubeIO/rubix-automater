package redis

import (
	"encoding/json"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/go-redis/redis/v8"
)

// CreateJobResult adds a new job result to the storage.
func (rs *Redis) CreateJobResult(result *model.JobResult) error {
	key := rs.getRedisKeyForJobResult(result.JobID)
	value, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetJobResult fetches a job result from the storage.
func (rs *Redis) GetJobResult(jobID string) (*model.JobResult, error) {
	key := rs.getRedisKeyForJobResult(jobID)
	val, err := rs.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, &apperrors.NotFoundErr{UUID: jobID, ResourceName: "job result"}
		}
		return nil, err
	}

	var result *model.JobResult
	err = json.Unmarshal(val, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateJobResult updates a job result to the storage.
func (rs *Redis) UpdateJobResult(jobID string, result *model.JobResult) error {
	key := rs.getRedisKeyForJobResult(jobID)
	value, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// DeleteJobResult deletes a job result from the storage.
func (rs *Redis) DeleteJobResult(jobID string) error {
	key := rs.getRedisKeyForJobResult(jobID)
	_, err := rs.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}
