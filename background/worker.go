package background

import (
	g "github.com/chiefsend/api/globals"
	"github.com/hibiken/asynq"
	"log"
)

func StartBackgroundWorkers() {
	r := asynq.RedisClientOpt{Addr: g.Conf.RedisAddr}
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
