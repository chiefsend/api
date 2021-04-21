package background

import (
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynq/inspeq"
	"log"
	"os"
	"strconv"
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
		workerss := os.Getenv("BACKGROUND_WORKERS")
		if workerss == "" {
			workers = 5
		} else {
			if workersi, err := strconv.Atoi(workerss); err != nil {
				log.Fatal(err)
			} else {
				workers = workersi
			}
		}
	}
	srv = asynq.NewServer(*redis, asynq.Config{
		Concurrency: workers,
	})
	// setup tasks
	mux := asynq.NewServeMux()
	mux.HandleFunc(DeleteShare, HandleDeleteShareTask)
	mux.HandleFunc(ShareEmail, HandleShareEmailTask)
	mux.HandleFunc(ContinuousDelete, HandleContinuousDeleteTask)
	// run server
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}

func StopBackgroundWorkers() {
	srv.Stop()
	//signalChan := make(chan os.Signal, 1)
	//signal.Notify(signalChan, os.Interrupt, os.Kill)
	//<-signalChan
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
