package main

import (
	"github.com/hibiken/asynq"
	"log"
)

func StartBackgroundWorker() {
	r := asynq.RedisClientOpt{Addr: "localhost:6379"}
	srv := asynq.NewServer(r, asynq.Config{
		Concurrency: 10,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(DeleteShare, HandleDeleteShareTask)
	mux.HandleFunc(ShareEmail, HandleShareEmailTask)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
