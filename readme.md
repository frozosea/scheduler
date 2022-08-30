## Scheduler

Scheduler is simple lib for golang for scheduled tasks.

Demonstration of functionality

```go
package main

import (
	"context"
	"fmt"
	"github.com/frozosea/scheduler/pkg"
	"log"
	"time"
)

func main() {
	taskManager := scheduler.NewDefault()
	ctx := context.Background()

	exampleFunc := func(ctx context.Context, exampleArgsSequence ...interface{}) scheduler.ShouldBeCancelled {
		for _, arg := range exampleArgsSequence {
			fmt.Println(arg)
		}
		return false
	}
	
	const customTaskId = "taskId"
	//time in format HH:MM, if put time not in this format 
	//executor will return timeParseError
	taskWithoutDuration, err := taskManager.Add(ctx, customTaskId, exampleFunc, "14:00", "argument 1", 2, "argument 3")
	if err != nil {
		log.Fatalf(`add task with string time error: %s`, err.Error())
	}
	fmt.Println(taskWithoutDuration.NextRunTime)

	
	
	_, err = taskManager.AddWithDuration(ctx, customTaskId+"2", exampleFunc, time.Second*4, "argument 1", 2, "argument 3")
	if err != nil {
		log.Fatalf(`add task with duration error: %s`, err.Error())
	}

	
	
	const newTime = "15:00"
	newTask, err := taskManager.Reschedule(ctx, customTaskId, newTime)
	if err != nil {
		log.Fatalf(`reschedule error: %s`, err.Error())
	}
	fmt.Println(newTask.NextRunTime)

	
	
	//return job instance 
	_, err = taskManager.RescheduleWithDuration(ctx, customTaskId, time.Hour)
	if err != nil {
		log.Fatalf(`reschedule with duration error: %s`, err.Error())
	}

	
	
	err = taskManager.Modify(ctx, customTaskId, exampleFunc, "new first arg", "new second args")
	if err != nil {
		log.Fatalf(`modify task error: %s`, err.Error())
	}

	
	
	//return job instance 
	_, err = taskManager.Get(ctx, customTaskId)
	if err != nil {
		log.Fatalf(`get job error: %s`, err.Error())
	}

	
	
	err = taskManager.Remove(ctx, customTaskId)
	if err != nil {
		log.Fatalf(`delete job error: %s`, err.Error())
	}

	
	
	err = taskManager.RemoveAll(ctx)
	if err != nil {
		log.Fatalf(`delete all jobs error: %s`, err.Error())
	}
}

````

