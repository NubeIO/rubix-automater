package redis

import (
	"encoding/json"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/timeconversion"
	"github.com/go-redis/redis/v8"
	"sort"
	"time"
)

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

	var jobs []*model.Job
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

// Recycle updates a job to the storage.
func (rs *Redis) Recycle(uuid string, j *model.Job) (*model.Job, error) {
	var err error

	var nextRunTime time.Time
	//if the job was completed and is enabled as cron
	if j.JobOptions != nil {
		if j.CompletedAt != nil {
			nextRunTime, err = timeconversion.AdjustTime(*j.CompletedAt, j.JobOptions.RunOnInterval)
			if err != nil {
				return nil, err
			}
			j.RunAt = &nextRunTime
		}
	}
	//now := ttime.New().Now()
	if j.RunAt != nil {
		//runAtTime, err := time.Parse(time.RFC3339Nano, body.RunAt.String())
		//if err != nil {
		//	return nil, &apperrors.ParseTimeErr{Message: err.Error()}
		//}
		//body.RunAt = &runAtTime
	} else {
		//body.RunAt = &now
	}

	j.Status = model.Pending
	j.ScheduledAt = nil
	j.StartedAt = nil

	err = rs.Watch(ctx, func(tx *redis.Tx) error {
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
	byJob, err := rs.GetTransactionsByJob(uuid)
	if err != nil {
		return err
	}
	for _, t := range byJob {
		err := rs.DeleteTransaction(t.UUID)
		if err != nil {
			return err
		}
	}
	_, err = rs.Del(ctx, key).Result()
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

	var dueJobs []*model.Job
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
