package commands

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"claude-company/internal/database"
	"claude-company/internal/models"
)

type AsyncTaskExecutor struct {
	taskRepo    *database.TaskRepository
	workers     chan struct{}
	taskQueue   chan string
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	runningTasks map[string]*models.Task
}

func NewAsyncTaskExecutor(maxWorkers int) *AsyncTaskExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AsyncTaskExecutor{
		taskRepo:     database.NewTaskRepository(),
		workers:      make(chan struct{}, maxWorkers),
		taskQueue:    make(chan string, 100),
		ctx:          ctx,
		cancel:       cancel,
		runningTasks: make(map[string]*models.Task),
	}
}

func (e *AsyncTaskExecutor) Start() {
	go e.processTaskQueue()
	go e.monitorPendingTasks()
	log.Println("Async task executor started")
}

func (e *AsyncTaskExecutor) Stop() {
	e.cancel()
	close(e.taskQueue)
	e.wg.Wait()
	log.Println("Async task executor stopped")
}

func (e *AsyncTaskExecutor) SubmitTask(taskID string) error {
	select {
	case e.taskQueue <- taskID:
		log.Printf("Task %s queued for execution", taskID)
		return nil
	case <-e.ctx.Done():
		return fmt.Errorf("executor is shutting down")
	default:
		return fmt.Errorf("task queue is full")
	}
}

func (e *AsyncTaskExecutor) processTaskQueue() {
	for {
		select {
		case taskID := <-e.taskQueue:
			if taskID == "" {
				return
			}
			
			select {
			case e.workers <- struct{}{}:
				e.wg.Add(1)
				go e.executeTask(taskID)
			case <-e.ctx.Done():
				return
			}
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *AsyncTaskExecutor) executeTask(taskID string) {
	defer func() {
		<-e.workers
		e.wg.Done()
		e.mu.Lock()
		delete(e.runningTasks, taskID)
		e.mu.Unlock()
	}()

	task, err := e.taskRepo.GetByID(taskID)
	if err != nil {
		log.Printf("Failed to get task %s: %v", taskID, err)
		return
	}

	e.mu.Lock()
	e.runningTasks[taskID] = task
	e.mu.Unlock()

	log.Printf("Starting execution of task %s: %s", taskID, task.Description)

	if err := e.taskRepo.UpdateStatus(taskID, "running"); err != nil {
		log.Printf("Failed to update task status to running: %v", err)
		return
	}

	err = e.performTaskExecution(task)
	
	if err != nil {
		log.Printf("Task %s failed: %v", taskID, err)
		if updateErr := e.taskRepo.UpdateStatus(taskID, "failed"); updateErr != nil {
			log.Printf("Failed to update task status to failed: %v", updateErr)
		}
		
		task.Result = fmt.Sprintf("Error: %v", err)
		task.UpdatedAt = time.Now()
		if updateErr := e.taskRepo.Update(task); updateErr != nil {
			log.Printf("Failed to update task result: %v", updateErr)
		}
	} else {
		log.Printf("Task %s completed successfully", taskID)
		if updateErr := e.taskRepo.UpdateStatus(taskID, "completed"); updateErr != nil {
			log.Printf("Failed to update task status to completed: %v", updateErr)
		}
	}
}

func (e *AsyncTaskExecutor) performTaskExecution(task *models.Task) error {
	switch task.Mode {
	case "ai":
		return e.executeAITask(task)
	case "manual":
		return e.executeManualTask(task)
	case "automated":
		return e.executeAutomatedTask(task)
	default:
		return fmt.Errorf("unknown task mode: %s", task.Mode)
	}
}

func (e *AsyncTaskExecutor) executeAITask(task *models.Task) error {
	log.Printf("Executing AI task: %s", task.Description)
	
	time.Sleep(time.Second * 2)
	
	task.Result = fmt.Sprintf("AI task completed: %s", task.Description)
	task.UpdatedAt = time.Now()
	return e.taskRepo.Update(task)
}

func (e *AsyncTaskExecutor) executeManualTask(task *models.Task) error {
	log.Printf("Manual task queued: %s", task.Description)
	
	task.Result = fmt.Sprintf("Manual task ready for execution: %s", task.Description)
	task.UpdatedAt = time.Now()
	return e.taskRepo.Update(task)
}

func (e *AsyncTaskExecutor) executeAutomatedTask(task *models.Task) error {
	log.Printf("Executing automated task: %s", task.Description)
	
	time.Sleep(time.Millisecond * 500)
	
	task.Result = fmt.Sprintf("Automated task completed: %s", task.Description)
	task.UpdatedAt = time.Now()
	return e.taskRepo.Update(task)
}

func (e *AsyncTaskExecutor) monitorPendingTasks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.processPendingTasks()
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *AsyncTaskExecutor) processPendingTasks() {
	tasks, err := e.taskRepo.GetByStatus("pending")
	if err != nil {
		log.Printf("Failed to get pending tasks: %v", err)
		return
	}

	for _, task := range tasks {
		select {
		case e.taskQueue <- task.ID:
			log.Printf("Auto-queued pending task: %s", task.ID)
		default:
			log.Printf("Task queue full, skipping task: %s", task.ID)
		}
	}
}

func (e *AsyncTaskExecutor) GetRunningTasks() map[string]*models.Task {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	result := make(map[string]*models.Task)
	for k, v := range e.runningTasks {
		result[k] = v
	}
	return result
}

func (e *AsyncTaskExecutor) GetQueueLength() int {
	return len(e.taskQueue)
}

func (e *AsyncTaskExecutor) GetActiveWorkers() int {
	return len(e.workers)
}