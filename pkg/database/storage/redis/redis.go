package redis

import (
	"context"
	"encoding/json"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"

	rs "github.com/NubeIO/rubix-automater/pkg/redis"
)

var _ automater.Storage = &Redis{}
var ctx = context.Background()

// Redis represents a redis client.
type Redis struct {
	*rs.Client
}

// New returns a redis client.
func New(url string, poolSize, minIdleConns int, keyPrefix string) *Redis {
	client := rs.New(url, poolSize, minIdleConns, keyPrefix, nil)
	return &Redis{
		Client: client,
	}
}

// CreateJob adds a new job to the storage.
func (rs *Redis) CreateJob(j *model.Job) error {
	key := rs.getRedisKeyForJob(j.UUID)
	value, err := json.Marshal(j)
	if err != nil {
		return err
	}

	err = rs.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetJob fetches a job from the storage.
func (rs *Redis) GetJob(uuid string) (*model.Job, error) {
	key := rs.getRedisKeyForJob(uuid)
	val, err := rs.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, &apperrors.NotFoundErr{UUID: uuid, ResourceName: "job"}
		}
		return nil, err
	}

	var j *model.Job
	err = json.Unmarshal(val, &j)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// GetJobs fetches all jobs from the storage, optionally filters the jobs by status.
func (rs *Redis) GetJobs(status model.JobStatus) ([]*model.Job, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey("job:*")
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	jobs := []*model.Job{}
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		j := &model.Job{}
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

// GetJobsByPipelineID fetches the jobs of the specified pipeline.
func (rs *Redis) GetJobsByPipelineID(pipelineID string) ([]*model.Job, error) {
	key := rs.getRedisKeyForPipeline(pipelineID)
	val, err := rs.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Mimic the relational storages behavior.
			return []*model.Job{}, nil
		}
		return nil, err
	}

	var p *model.Pipeline
	err = json.Unmarshal(val, &p)
	if err != nil {
		return nil, err
	}

	return p.Jobs, nil
}

// UpdateJob updates a job to the storage.
func (rs *Redis) UpdateJob(uuid string, j *model.Job) (*model.Job, error) {
	err := rs.Watch(ctx, func(tx *redis.Tx) error {
		key := rs.getRedisKeyForJob(uuid)
		value, err := json.Marshal(j)
		if err != nil {
			return err
		}

		err = rs.Set(ctx, key, value, 0).Err()
		if err != nil {
			return err
		}
		if j.BelongsToPipeline() {
			// Sync pipeline job.
			pipelineKey := rs.getRedisKeyForPipeline(j.PipelineID)
			val, err := rs.Get(ctx, pipelineKey).Bytes()
			if err != nil {
				return err
			}

			var p *model.Pipeline
			err = json.Unmarshal(val, &p)
			if err != nil {
				return err
			}
			for i, job := range p.Jobs {
				if job.UUID == j.UUID {
					p.Jobs[i] = j
				}
			}
			err = rs.UpdatePipeline(p.UUID, p)
			if err != nil {
				return err
			}
		}
		return nil
	})
	job, err := rs.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// DeleteJob deletes a job from the storage.
func (rs *Redis) DeleteJob(uuid string) error {
	key := rs.getRedisKeyForJob(uuid)
	_, err := rs.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (rs *Redis) GetDueJobs() ([]*model.Job, error) {
	var keys []string
	key := rs.GetRedisPrefixedKey("job:*")
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	dueJobs := []*model.Job{}
	for _, key := range keys {
		value, err := rs.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		j := &model.Job{}
		if err := json.Unmarshal(value, j); err != nil {
			return nil, err
		}
		if j.IsScheduled() {
			if j.RunAt.Before(time.Now()) && j.Status == model.Pending {
				dueJobs = append(dueJobs, j)
			}
		}
	}

	// ORDER BY run_at ASC
	sort.Slice(dueJobs, func(i, j int) bool {
		return dueJobs[i].RunAt.Before(*dueJobs[j].RunAt)
	})
	return dueJobs, nil
}

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
	key := rs.GetRedisPrefixedKey("pipeline:*")
	iter := rs.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	pipelines := []*model.Pipeline{}
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

		jobs := []*model.Job{}
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

func (rs *Redis) getRedisKeyForPipeline(pipelineID string) string {
	return rs.GetRedisPrefixedKey("pipeline:" + pipelineID)
}

func (rs *Redis) getRedisKeyForJob(jobID string) string {
	return rs.GetRedisPrefixedKey("job:" + jobID)
}

func (rs *Redis) getRedisKeyForJobResult(jobID string) string {
	return rs.GetRedisPrefixedKey("jobresult:" + jobID)
}
