# rubix-automater

## install
```
go mod tidy
```
```
go run ./cmd/cmd main.go
```

### install redis
https://redis.io/docs/getting-started/installation/install-redis-on-linux/

### redis cli

#### wipe the db
https://redis.io/docs/getting-started/

this will clear the db

```
redis-cli flushdb
```

### Job

The implementation uses the notion of `job`, which describes the work that needs to be done and carries information about the task that will run for the
specific job. User-defined tasks are assigned to jobs. Every job can be assigned with a different task, a JSON payload with the data required for the task
to be executed, and an optional timeout interval. Jobs can be scheduled to run at a specified time or instantly.

After the tasks have been executed, their results along with the errors (if any) are stored into a storage system.

### Pipeline

A `pipeline` is a sequence of jobs that need to be executed in a specified order, one by one. Every job in the pipeline can be assigned with a different task
and parameters, and each task callback can optionally use the results of the previous task in the job sequence. A pipeline can also be scheduled to run sometime in the future, or immediately.



## Usage

Create a new job by making a `/jobs` service method. You can inject arbitrary parameters for your task to run
by including them in the request body.

- POST

```json
{
  "name": "run ping",
  "description": "run a ping",
  "task_name": "pinghost",
  "task_params": {
    "url": "nube-io.com",
    "port": 443,
    "errorOnFailSetting": 10
  },
  "timeout_in_sec": 100
}
```

To schedule a new job to run at a specific time, add `run_at` field to the request body.

- POST

```json
{
  "name": "run ping",
  "description": "run a ping",
  "task_name": "pinghost",
  "run_at": "2022-06-04T02:23:00.426752075+10:00",
  "task_params": {
    "url": "nube-io.com",
    "port": 443,
    "errorOnFailSetting": 10
  },
  "timeout_in_sec": 100
}
```

Create a new pipeline by making a POST HTTP call to `/pipelines` service method. You can inject arbitrary parameters
for your tasks to run by including them in the request body. Optionally, you can tune your tasks to use any results of the previous task in the pipeline, creating
a `bash`-like command pipeline. Pipelines can also be scheduled for execution at a specific time, by adding `run_at` field to the request payload
just like it's done with the jobs.


The example below will in steps

- try and ping the host (if ping failed it will skip the installation app, but only if this is true `use_previous_results`)
- if it found the host install a new app


if `run_at"` is `nil` then it will just run the job once

```json
{
  "name": "instal a rubix app",
  "description": "first testing if the device is offline then try and install the app",
  "jobs": [
    {
      "name": "pingHost",
      "description": "host ping",
      "task_name": "pingHost",
      "task_params": {
        "url": "0.0.0.0",
        "port": 8090,
        "errorOnFailSetting": 10
      }
    },
    {
      "name": "installApp",
      "description": "try and install a rubix-app",
      "task_name": "installApp",
      "task_params": {
        "url": "0.0.0.0",
        "port": 8090,
        "AppName": "flow-framework",
        "version": "latest",
        "token": "GIT TOKEN"
      },
      "use_previous_results": true,
      "timeout": 120
    }
  ]
}
```