package scheduler

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/rakanalh/scheduler/storage"
	"github.com/rakanalh/scheduler/task"
	"github.com/rs/zerolog/log"
)

type WasteFunc = interface{}

type wasteScheduler struct {
	taskFunctions map[string]WasteFunc
	funcRegistry  *task.FuncRegistry
	stopChan      chan bool
	tasks         map[task.TaskID]*task.Task
	store         storage.TaskStore
}

var WS *wasteScheduler

// root@localhost(db-mysql-others-local0a.42):[information_schema]> select TABLE_SCHEMA,TABLE_NAME  from information_schema.TABLES where TABLE_NAME like '__waste_%';
// +--------------+----------------------------+
// | TABLE_SCHEMA | TABLE_NAME                 |
// +--------------+----------------------------+
// | greyhound_db | __waste_2019_2_28_jonyTest |
// +--------------+----------------------------+
// 1 row in set (0.00 sec)

func SetupScheduler() *wasteScheduler {
	storage := storage.NewSqlite3Storage(
		storage.Sqlite3Config{
			DbName: "waste.db",
		},
	)
	if err := storage.Connect(); err != nil {
		log.Fatal().Err(err).Msg("Failed to get bindings from  ")
	}

	if err := storage.Initialize(); err != nil {
		log.Fatal().Err(err).Msg("Failed to get bindings from  ")
	}

	funcRegistry := task.NewFuncRegistry()

	w := wasteScheduler{
		taskFunctions: make(map[string]WasteFunc),
		funcRegistry:  funcRegistry,
		stopChan:      make(chan bool),
		tasks:         make(map[task.TaskID]*task.Task),
		store:         storage,
	}

	return &w
}

func (scheduler *wasteScheduler) Stop() {
	scheduler.stopChan <- true
}

func (scheduler *wasteScheduler) Start() error {
	log.Info().Msg("Scheduler is starting...")

	// Populate tasks from storage
	if err := scheduler.populateTasks(); err != nil {
		return nil
	}

	scheduler.runPending()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ticker.C:
				scheduler.runPending()
			case <-scheduler.stopChan:
				close(scheduler.stopChan)
			}
		}
	}()

	return nil
}

func (ws *wasteScheduler) RegisterTask(task string, function task.Function) {
	ws.taskFunctions[task] = function
	ws.funcRegistry.Add(function)

}

func (ws *wasteScheduler) AddTask(days int, task string, arg interface{}) {
	taskId, err := ws.runAt(time.Now().AddDate(0, 0, days), ws.taskFunctions[task], arg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to Add Task ")
	}
	log.Info().Msgf("stored task to %s, in %d days, ID: %s", task, days, taskId)
}

func (scheduler *wasteScheduler) runAt(time time.Time, function task.Function, params ...task.Param) (task.TaskID, error) {
	funcMeta, err := scheduler.funcRegistry.Add(function)
	if err != nil {
		return "", err
	}

	task := task.New(funcMeta, params)

	task.NextRun = time

	scheduler.tasks[task.Hash()] = task
	taskAttr, err := getTaskAttributes(task)
	if err != nil {
		return "", err
	}
	err = scheduler.store.Add(taskAttr)
	if err != nil {
		return "", err
	}
	return task.Hash(), nil
}

func (scheduler *wasteScheduler) persistRegisteredTasks() error {
	for _, task := range scheduler.tasks {
		taskAttr, err := getTaskAttributes(task)
		if err != nil {
			return err
		}
		err = scheduler.store.Add(taskAttr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (scheduler *wasteScheduler) GetTasks() (map[task.TaskID]*task.Task, error) {
	return scheduler.tasks, nil
}

func (scheduler *wasteScheduler) runPending() {
	for _, task := range scheduler.tasks {
		if task.IsDue() {
			go task.Run()

			if !task.IsRecurring {
				taskAttr, err := getTaskAttributes(task)
				if err != nil {
					return
				}
				_ = scheduler.store.Remove(taskAttr)
				delete(scheduler.tasks, task.Hash())
			}
		}
	}
}

func (scheduler *wasteScheduler) populateTasks() error {
	tasks, err := scheduler.store.Fetch()
	if err != nil {
		return err
	}

	for _, dbTask := range tasks {

		// If the task instance is still registered with the same computed hash then move on.
		// Otherwise, add the task
		// @todo: Fix this - fetch doesn't bring back the Hash - this compare will always fail...
		_, ok := scheduler.tasks[task.TaskID(dbTask.Hash)]
		if !ok {
			log.Printf("Detected a change in attributes of one of the instances of task %s, \n",
				dbTask.Name)
			funcMeta, _ := scheduler.funcRegistry.Get(dbTask.Name)
			params, _ := paramsFromString(funcMeta, dbTask.Params)

			nextRun, _ := time.Parse(time.RFC3339, dbTask.NextRun)
			lastRun, _ := time.Parse(time.RFC3339, dbTask.LastRun)
			duration, _ := time.ParseDuration(dbTask.Duration)
			isRecurring, _ := strconv.Atoi(dbTask.IsRecurring)

			registerTask := task.NewWithSchedule(funcMeta, params, task.Schedule{
				IsRecurring: isRecurring == 1,
				Duration:    time.Duration(duration),
				LastRun:     lastRun,
				NextRun:     nextRun,
			})
			scheduler.tasks[registerTask.Hash()] = registerTask
		}

	}
	return nil
}

func getTaskAttributes(task *task.Task) (storage.TaskAttributes, error) {
	params, err := paramsToString(task.Params)
	if err != nil {
		return storage.TaskAttributes{}, err
	}

	isRecurring := 0
	if task.IsRecurring {
		isRecurring = 1
	}

	return storage.TaskAttributes{
		Hash:        string(task.Hash()),
		Name:        task.Func.Name,
		LastRun:     task.LastRun.Format(time.RFC3339),
		NextRun:     task.NextRun.Format(time.RFC3339),
		Duration:    task.Duration.String(),
		IsRecurring: strconv.Itoa(isRecurring),
		Params:      params,
	}, nil
}

func paramsToString(params []task.Param) (string, error) {
	var paramsList []string
	for _, param := range params {
		paramStr, err := json.Marshal(param)
		if err != nil {
			return "", err
		}
		paramsList = append(paramsList, string(paramStr))
	}
	data, err := json.Marshal(paramsList)
	return string(data), err
}

func paramsFromString(funcMeta task.FunctionMeta, payload string) ([]task.Param, error) {
	var params []task.Param
	if strings.TrimSpace(payload) == "" {
		return params, nil
	}
	paramTypes := funcMeta.Params()
	var paramsStrings []string
	err := json.Unmarshal([]byte(payload), &paramsStrings)
	if err != nil {
		return params, err
	}
	for i, paramStr := range paramsStrings {
		paramType := paramTypes[i]
		target := reflect.New(paramType)
		err := json.Unmarshal([]byte(paramStr), target.Interface())
		if err != nil {
			return params, err
		}
		param := reflect.Indirect(target).Interface().(task.Param)
		params = append(params, param)
	}

	return params, nil
}
