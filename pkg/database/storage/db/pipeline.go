package redis

import (
	"encoding/json"
	"fmt"
	pprint "github.com/NubeIO/edge/pkg/helpers/print"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/go-redis/redis/v8"
	"sort"
	"time"
)

// CreatePipeline adds a new pipeline and of its jobs to the storage.
func (inst *Redis) CreatePipeline(p *model.Pipeline) error {
	pprint.PrintJOSN(p)
	err := inst.Watch(ctx, func(tx *redis.Tx) error {

		for _, j := range p.Jobs {
			key := inst.getRedisKeyForJob(j.UUID)
			value, err := json.Marshal(j)
			if err != nil {
				return err
			}
			err = inst.Set(ctx, key, value, 0).Err()
			if err != nil {
				return err
			}
		}
		key := inst.getRedisKeyForPipeline(p.UUID)
		value, err := json.Marshal(p)
		if err != nil {
			return err
		}
		err = inst.Set(ctx, key, value, 0).Err()
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
func (inst *Redis) GetPipeline(uuid string) (*model.Pipeline, error) {
	key := inst.getRedisKeyForPipeline(uuid)
	val, err := inst.Get(ctx, key).Bytes()
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
func (inst *Redis) GetPipelines(status model.JobStatus) ([]*model.Pipeline, error) {
	var keys []string
	key := inst.GetRedisPrefixedKey(fmt.Sprintf("%s:*", pipeline))
	iter := inst.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	var pipelines []*model.Pipeline
	for _, key := range keys {
		value, err := inst.Get(ctx, key).Bytes()
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

// RecyclePipeline updates a pipeline to the storage.
func (inst *Redis) RecyclePipeline(uuid string, p *model.Pipeline) (*model.Pipeline, error) {
	getExisting := p
	jobs, err := inst.GetJobsByPipelineID(uuid) //get the existing pipeline jobs
	if err != nil {
		return nil, err
	}

	var nextRunTime = time.Time{}
	var recycleJobs []*model.Job
	for i, job := range jobs {
		now, err := automater.PipelineRunAt(p.PipelineOptions.RunOnInterval, p.PipelineOptions, i)
		runAtTime := now.Add(time.Millisecond * time.Duration(i+2)) // in db GetDueJobs it orders by time desc, so we need a small buffer (this is a hack)
		if err != nil {
			return nil, err
		}
		job.RunAt = &runAtTime
		recycleJob, err := inst.Recycle(job.UUID, job) // recycle jobs
		if err != nil {
			return nil, err
		}
		recycleJob.RunAt = &nextRunTime
		recycleJobs = append(recycleJobs, recycleJob)
	}

	getExisting.Jobs = recycleJobs
	getExisting.Status = model.Pending
	getExisting.RunAt = &nextRunTime
	getExisting.StartedAt = nil
	getExisting.Duration = nil

	err = inst.UpdatePipeline(uuid, getExisting)
	if err != nil {
		return nil, err
	}
	getPipeline, err := inst.GetPipeline(uuid)
	if err != nil {
		return nil, err
	}
	return getPipeline, err
}

// UpdatePipeline updates a pipeline to the storage.
func (inst *Redis) UpdatePipeline(uuid string, p *model.Pipeline) error {
	key := inst.getRedisKeyForPipeline(uuid)
	value, err := json.Marshal(p)
	if err != nil {
		return err
	}

	err = inst.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// DeletePipeline deletes a pipeline and all its jobs from the storage.
func (inst *Redis) DeletePipeline(uuid string) error {
	err := inst.Watch(ctx, func(tx *redis.Tx) error {
		var keys []string
		key := inst.GetRedisPrefixedKey("job:*")
		iter := inst.Scan(ctx, 0, key, 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return err
		}

		var jobs []*model.Job
		for _, key := range keys {
			value, err := inst.Get(ctx, key).Bytes()
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
			key = inst.getRedisKeyForJobResult(j.UUID)
			_, err := inst.Del(ctx, key).Result()
			if err != nil {
				return err
			}
			key = inst.getRedisKeyForJob(j.UUID)
			_, err = inst.Del(ctx, key).Result()
			if err != nil {
				return err
			}
		}
		key = inst.getRedisKeyForPipeline(uuid)
		_, err := inst.Del(ctx, key).Result()
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
