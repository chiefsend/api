package background

import (
	"github.com/hibiken/asynq"
	"log"
	"os"
	"time"
)

var redis *asynq.RedisClientOpt

func StartBackgroundWorkers() {
	redis = &asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URI")}
	srv := asynq.NewServer(*redis, asynq.Config{
		Concurrency: 10,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(DeleteShare, HandleDeleteShareTask)
	mux.HandleFunc(ShareEmail, HandleShareEmailTask)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}

func EnqueueJob(task *asynq.Task, at *time.Time) error {
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
