package main

import (
	"time"

	"github.com/outbrain/golib/log"

	"github.com/cohenjo/waste/go/http"
	"github.com/rakanalh/scheduler"
	"github.com/rakanalh/scheduler/storage"
)

func TaskWithoutArgs() {
	log.Infof("TaskWithoutArgs is executed")
}

func TaskWithArgs(message string) {
	log.Infof("TaskWithArgs is executed. message:", message)
}

func main() {
	log.SetLevel(log.INFO)
	storage := storage.NewSqlite3Storage(
		storage.Sqlite3Config{
			DbName: "task_store.db",
		},
	)
	if err := storage.Connect(); err != nil {
		log.Critical("Could not connect to db", err)
	}

	if err := storage.Initialize(); err != nil {
		log.Critical("Could not intialize database", err)
	}

	s := scheduler.New(storage)

	// Start a task without arguments
	if _, err := s.RunAfter(30*time.Second, TaskWithoutArgs); err != nil {
		log.Criticale(err)
	}

	// Start a task with arguments
	// if _, err := s.RunEvery(5*time.Second, TaskWithArgs, "Hello from recurring task 1"); err != nil {
	// 	log.Criticale(err)
	// }

	if _, err := s.RunAt(time.Now().Add(2*time.Minute), TaskWithArgs, "Hello from recurring task 2"); err != nil {
		log.Criticale(err)
	}
	// Start the same task as above with a different argument
	// if _, err := s.RunEvery(10*time.Second, TaskWithArgs, "Hello from recurring task 2"); err != nil {
	// 	log.Criticale(err)
	// }
	tasks, _ := storage.Fetch()
	log.Infof("stored tasks: \n%v\n", tasks)
	s.Start()
	http.Serve()
	s.Wait()
}
