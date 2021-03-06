package redis

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"github.com/NubeIO/rubix-automater/pkg/helpers/ttime"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"math"
	"sort"
	"time"
)

// CreateJob adds a new job to the storage.
func (inst *Redis) CreateJob(j *model.Job) error {
	key := inst.getRedisKeyForJob(j.UUID)
	value, err := json.Marshal(j)
	if err != nil {
		return err
	}
	err = inst.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetJob fetches a job from the storage.
func (inst *Redis) GetJob(uuid string) (*model.Job, error) {
	key := inst.getRedisKeyForJob(uuid)
	val, err := inst.Get(ctx, key).Bytes()
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
func (inst *Redis) GetJobs(status model.JobStatus) ([]*model.Job, error) {
	var keys []string
	key := inst.GetRedisPrefixedKey("job:*")
	iter := inst.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	var jobs []*model.Job
	for _, key := range keys {
		value, err := inst.Get(ctx, key).Bytes()
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
func (inst *Redis) GetJobsByPipelineID(pipelineID string) ([]*model.Job, error) {
	key := inst.getRedisKeyForPipeline(pipelineID)
	val, err := inst.Get(ctx, key).Bytes()
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
func (inst *Redis) Recycle(uuid string, j *model.Job) (*model.Job, error) {
	var err error
	_, err = inst.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	now := ttime.New().Now()
	j.UUID = uuid
	j.Status = model.Pending
	j.ScheduledAt = nil
	j.StartedAt = nil
	j.CreatedAt = &now
	err = inst.Watch(ctx, func(tx *redis.Tx) error {
		key := inst.getRedisKeyForJob(uuid)
		value, err := json.Marshal(j)
		if err != nil {
			return err
		}
		err = inst.Set(ctx, key, value, 0).Err()
		if err != nil {
			return err
		}
		return nil
	})
	job, err := inst.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// UpdateJob updates a job to the storage.
func (inst *Redis) UpdateJob(uuid string, j *model.Job) (*model.Job, error) {
	err := inst.Watch(ctx, func(tx *redis.Tx) error {
		key := inst.getRedisKeyForJob(uuid)
		value, err := json.Marshal(j)
		if err != nil {
			return err
		}

		err = inst.Set(ctx, key, value, 0).Err()
		if err != nil {
			return err
		}
		if j.BelongsToPipeline() {
			// Sync pipeline job.
			pipelineKey := inst.getRedisKeyForPipeline(j.PipelineID)
			val, err := inst.Get(ctx, pipelineKey).Bytes()
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
			err = inst.UpdatePipeline(p.UUID, p)
			if err != nil {
				return err
			}
		}
		return nil
	})
	job, err := inst.GetJob(uuid)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// DeleteJob deletes a job from the storage.
func (inst *Redis) DeleteJob(uuid string) error {
	key := inst.getRedisKeyForJob(uuid)
	byJob, err := inst.GetTransactionsByJob(uuid)
	if err != nil {
		return err
	}
	for _, t := range byJob {
		err := inst.DeleteTransaction(t.UUID)
		if err != nil {
			return err
		}
	}
	_, err = inst.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

func timeDif(t1, t2 time.Time) string {
	hs := t1.Sub(t2).Hours()
	hs, mf := math.Modf(hs)
	ms := mf * 60
	ms, sf := math.Modf(ms)
	ss := sf * 60
	return fmt.Sprintf("%f hours %f minutes %f seconds", hs, ms, ss)

}

// GetDueJobs fetches all jobs scheduled to run before now and have not been scheduled yet.
func (inst *Redis) GetDueJobs() ([]*model.Job, error) {
	var keys []string
	key := inst.GetRedisPrefixedKey("job:*")
	iter := inst.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	var dueJobs []*model.Job
	var pendingJobs []*model.Job
	for _, key := range keys {

		value, err := inst.Get(ctx, key).Bytes()
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
			if j.RunAt.After(time.Now()) && j.Status == model.Pending {
				pendingJobs = append(pendingJobs, j)
				logrus.Infof("name:%s task:%s, will run at:%s", j.Name, j.TaskName, timeDif(*j.RunAt, time.Now()))

			}

		}
	}
	logrus.Infof("get due jobs count:%d due", len(pendingJobs))
	// ORDER BY run_at ASC
	sort.Slice(dueJobs, func(i, j int) bool {
		return dueJobs[i].RunAt.Before(*dueJobs[j].RunAt)
	})
	return dueJobs, nil
}
