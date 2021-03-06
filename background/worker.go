package background

import (
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynq/inspeq"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var redis *asynq.RedisClientOpt
var srv *asynq.Server

func StartBackgroundWorkers() {
	// create config
	var db int
	{
		dbs := os.Getenv("REDIS_DB")
		if dbs == "" {
			db = 0
		} else {
			if dbi, err := strconv.Atoi(dbs); err != nil {
				log.Fatal(err)
			} else {
				db = dbi
			}
		}
	}
	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		redis = &asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URI"), DB: db}
	} else {
		redis = &asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URI"), DB: db, Password: password}
	}
	// create server
	var workers int
	{
		workerCount := os.Getenv("BACKGROUND_WORKERS")
		if workerCount == "" {
			workers = 5
		} else {
			if workersInt, err := strconv.Atoi(workerCount); err != nil {
				log.Fatal(err)
			} else {
				workers = workersInt
			}
		}
	}
	srv = asynq.NewServer(*redis, asynq.Config{
		Concurrency: workers,
	})
	// setup tasks
	mux := asynq.NewServeMux()
	mux.HandleFunc(DeleteShare, HandleDeleteShareTask)
	mux.HandleFunc(ContinuousDelete, HandleContinuousDeleteTask)
	// run server
	if err := srv.Start(mux); err != nil {
		log.Fatal(err)
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals
	StopBackgroundWorkers()
	fmt.Println("stopped background workers ...")
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}

func StopBackgroundWorkers() {
	srv.Stop()
}

func EnqueueJob(task *asynq.Task, at *time.Time) error {
	// workers need to be running
	if redis == nil {
		return nil
	}

	client := asynq.NewClient(*redis)
	if at != nil {
		if _, err := client.Enqueue(task, asynq.ProcessAt(*at)); err != nil {
			return err
		}
	} else {
		if _, err := client.Enqueue(task); err != nil {
			return err
		}
	}
	return nil
}

func GetJobs() ([]*inspeq.SchedulerEntry, error) {
	// get inspector
	i := inspeq.New(redis)
	defer i.Close()
	// get jobs
	jobs, err := i.SchedulerEntries()
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
