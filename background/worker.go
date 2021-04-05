package background

import (
	"github.com/hibiken/asynq"
	"log"
	"os"
)

func StartBackgroundWorkers() {
	r := asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URI")}
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
