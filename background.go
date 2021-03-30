package main

import (
	"fmt"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

// Make a redis pool
var redisPool = &redis.Pool{
	MaxActive: 5,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", ":6379")
	},
}

// Make an enqueuer with a particular namespace
var enqueuer = work.NewEnqueuer("ChiefSend", redisPool)

type Context struct {
	ShareID uuid.UUID
}

func ConfigurePool() {
	// Make a new pool. Arguments:
	// Context{} is a struct that will be the context for the request.
	// 10 is the max concurrency
	// "ChiefSend" is the Redis namespace
	pool := work.NewWorkerPool(Context{}, 10, "ChiefSend", redisPool)
	// Add middleware that will be executed for each job
	pool.Middleware((*Context).FindShare)
	pool.Middleware((*Context).Log)
	pool.Job("DeleteShare", (*Context).DeleteShare)
	pool.Start()
	//signalChan := make(chan os.Signal, 1)
	//signal.Notify(signalChan, os.Interrupt, os.Kill)
	//<-signalChan
	//pool.Stop()
}

func (c *Context) FindShare(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Println("FindShare (background job")
	if _, ok := job.Args["ShareID"]; ok {
		c.ShareID = uuid.MustParse(job.ArgString("ShareID"))
		if err := job.ArgError(); err != nil {
			return err
		}
	}
	return next()
}

func (c *Context) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Printf("Starting job: %s (%s)", job.Name, c.ShareID)
	return next()
}

func (c *Context) DeleteShare(job *work.Job) error {
	return nil // TODO
}
