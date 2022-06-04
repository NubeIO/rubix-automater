package redis

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/go-redis/redis/v8"
	"sort"
)

// CreatePipeline adds a new pipeline and of its jobs to the storage.
func (rs *Redis) CreatePipeline(p *model.Pipeline) error {
	err := rs.Watch(ctx, func(tx *redis.Tx) error {

		for _, j := range p.Jobs {
			key := rs.getRedisKeyForJob(j.UUID)
			value, err := json.Marshal(j)
			if err != nil {
				return err
			}

			err = rs.Set(ctx, key, value, 0).Err()
			if err != nil {
				return err
			}
		}

		key := rs.getRedisKeyForPipeline(p.UUID)
		value, err := json.Marshal(p)
		if err != nil {
			return err
		}

		err = rs.Set(ctx, key, value, 0).Err()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetPipeline fetches a pipeline from the storage.
func (rs *Redis) GetPipeline(uuid string) (*model.Pipeline, error) {
	key := rs.getRedisKeyForPipeline(uuid)
	val, err := rs.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, &apperrors.NotFoundErr{UUID: uuid, ResourceName: "pipeline"}
		}
		return nil, err
	}

	var p *model.Pipeline
	err = json.Unmarshal(val, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetPipelines fetches all pipelines from the storage, optionally filters the pipelines by status.
func (rs *Redis) GetPipelines(status model.JobStatus) ([]*model.Pipeline, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey(fmt.Sprintf("%s:*", pipeline))
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	var pipelines []*model.Pipeline
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		p := &model.Pipeline{}
		if err := json.Unmarshal(value, p); err != nil {
			return nil, err
		}
		if status == model.Undefined || p.Status == status {
			pipelines = append(pipelines, p)
		}
	}

	// ORDER BY created_at ASC
	sort.Slice(pipelines, func(i, j int) bool {
		return pipelines[i].CreatedAt.Before(*pipelines[j].CreatedAt)
	})
	return pipelines, nil

}

// UpdatePipeline updates a pipeline to the storage.
func (rs *Redis) UpdatePipeline(uuid string, p *model.Pipeline) error {
	key := rs.getRedisKeyForPipeline(uuid)
	value, err := json.Marshal(p)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// DeletePipeline deletes a pipeline and all its jobs from the storage.
func (rs *Redis) DeletePipeline(uuid string) error {
	err := rs.Watch(ctx, func(tx *redis.Tx) error {
		var keys []string
		key := rs.GetRedisPrefixedKey("job:*")
		iter := rs.Scan(ctx, 0, key, 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return err
		}

		var jobs []*model.Job
		for _, key := range keys {
			value, err := rs.Get(ctx, key).Bytes()
			if err != nil {
				return err
			}
			j := &model.Job{}
			if err := json.Unmarshal(value, j); err != nil {
				return err
			}
			if j.PipelineID == uuid {
				jobs = append(jobs, j)
			}
		}
		for _, j := range jobs {
			key = rs.getRedisKeyForJobResult(j.UUID)
			_, err := rs.Del(ctx, key).Result()
			if err != nil {
				return err
			}
			key = rs.getRedisKeyForJob(j.UUID)
			_, err = rs.Del(ctx, key).Result()
			if err != nil {
				return err
			}
		}
		key = rs.getRedisKeyForPipeline(uuid)
		_, err := rs.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
