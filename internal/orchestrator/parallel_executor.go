package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ParallelExecutor struct {
	mu              sync.RWMutex
	config          ParallelExecutorConfig
	executionPool   *ExecutionPool
	activeJobs      map[string]*ExecutionJob
	eventBus        EventBus
	metrics         *ExecutorMetrics
}

type ParallelExecutorConfig struct {
	MaxConcurrentJobs    int           `json:"max_concurrent_jobs"`
	DefaultJobTimeout    time.Duration `json:"default_job_timeout"`
	RetryPolicy          RetryPolicy   `json:"retry_policy"`
	ResourceLimits       ResourceLimits `json:"resource_limits"`
	HealthCheckInterval  time.Duration `json:"health_check_interval"`
}

type ResourceLimits struct {
	MaxMemoryMB     int `json:"max_memory_mb"`
	MaxCPUPercent   int `json:"max_cpu_percent"`
	MaxDiskSpaceMB  int `json:"max_disk_space_mb"`
}

type ExecutionPool struct {
	workers       chan *Worker
	jobQueue      chan *ExecutionJob
	activeWorkers sync.Map
	shutdown      chan struct{}
	wg            sync.WaitGroup
}

type ExecutionJob struct {
	ID              string                    `json:"id"`
	Type            JobType                   `json:"type"`
	Task            *Task                     `json:"task,omitempty"`
	SubTask         *SubTask                  `json:"subtask,omitempty"`
	Step            *TaskStep                 `json:"step,omitempty"`
	Context         context.Context           `json:"-"`
	Cancel          context.CancelFunc        `json:"-"`
	StartTime       time.Time                 `json:"start_time"`
	EndTime         *time.Time                `json:"end_time,omitempty"`
	Status          JobStatus                 `json:"status"`
	Progress        float64                   `json:"progress"`
	Result          *ExecutionResult          `json:"result,omitempty"`
	Error           error                     `json:"-"`
	RetryCount      int                       `json:"retry_count"`
	AssignedWorker  *Worker                   `json:"assigned_worker,omitempty"`
	Dependencies    []string                  `json:"dependencies"`
	Executor        JobExecutorFunc           `json:"-"`
	ProgressCallback ProgressCallbackFunc     `json:"-"`
}

type JobType string

const (
	JobTypeTask    JobType = "task"
	JobTypeSubTask JobType = "subtask"
	JobTypeStep    JobType = "step"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusQueued     JobStatus = "queued"
	JobStatusRunning    JobStatus = "running"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
	JobStatusRetrying   JobStatus = "retrying"
)

type ExecutionResult struct {
	Success     bool              `json:"success"`
	Output      interface{}       `json:"output"`
	Error       *ExecutionError   `json:"error,omitempty"`
	Metadata    map[string]any    `json:"metadata"`
	Duration    time.Duration     `json:"duration"`
	ResourceUsage *ResourceUsage  `json:"resource_usage,omitempty"`
}

type ExecutionError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	Retryable  bool           `json:"retryable"`
	Timestamp  time.Time      `json:"timestamp"`
}

type ResourceUsage struct {
	PeakMemoryMB    int           `json:"peak_memory_mb"`
	AvgCPUPercent   float64       `json:"avg_cpu_percent"`
	DiskSpaceUsedMB int           `json:"disk_space_used_mb"`
	NetworkBytesIn  int64         `json:"network_bytes_in"`
	NetworkBytesOut int64         `json:"network_bytes_out"`
	Duration        time.Duration `json:"duration"`
}

type ExecutorMetrics struct {
	mu                    sync.RWMutex
	TotalJobsExecuted     int64         `json:"total_jobs_executed"`
	SuccessfulJobs        int64         `json:"successful_jobs"`
	FailedJobs            int64         `json:"failed_jobs"`
	CancelledJobs         int64         `json:"cancelled_jobs"`
	AvgExecutionTime      time.Duration `json:"avg_execution_time"`
	CurrentConcurrentJobs int           `json:"current_concurrent_jobs"`
	PeakConcurrentJobs    int           `json:"peak_concurrent_jobs"`
	LastUpdateTime        time.Time     `json:"last_update_time"`
}

type JobExecutorFunc func(ctx context.Context, job *ExecutionJob) (*ExecutionResult, error)
type ProgressCallbackFunc func(jobID string, progress float64)

func NewParallelExecutor(config ParallelExecutorConfig, eventBus EventBus) *ParallelExecutor {
	if config.MaxConcurrentJobs <= 0 {
		config.MaxConcurrentJobs = 10
	}
	if config.DefaultJobTimeout <= 0 {
		config.DefaultJobTimeout = 30 * time.Minute
	}
	if config.HealthCheckInterval <= 0 {
		config.HealthCheckInterval = 30 * time.Second
	}

	pe := &ParallelExecutor{
		config:      config,
		activeJobs:  make(map[string]*ExecutionJob),
		eventBus:    eventBus,
		metrics:     &ExecutorMetrics{LastUpdateTime: time.Now()},
	}

	pe.executionPool = pe.createExecutionPool()
	
	return pe
}

func (pe *ParallelExecutor) createExecutionPool() *ExecutionPool {
	pool := &ExecutionPool{
		workers:  make(chan *Worker, pe.config.MaxConcurrentJobs),
		jobQueue: make(chan *ExecutionJob, pe.config.MaxConcurrentJobs*2),
		shutdown: make(chan struct{}),
	}

	for i := 0; i < pe.config.MaxConcurrentJobs; i++ {
		worker := &Worker{
			ID:           fmt.Sprintf("worker_%d", i),
			Type:         "execution_worker",
			Status:       WorkerStatusIdle,
			Capabilities: []string{"task_execution", "step_execution"},
			LastSeen:     time.Now(),
		}
		pool.workers <- worker
	}

	pool.wg.Add(1)
	go pe.processJobs(pool)

	return pool
}

func (pe *ParallelExecutor) processJobs(pool *ExecutionPool) {
	defer pool.wg.Done()

	for {
		select {
		case job := <-pool.jobQueue:
			worker := <-pool.workers
			pool.wg.Add(1)
			go pe.executeJob(job, worker, pool)

		case <-pool.shutdown:
			return
		}
	}
}

func (pe *ParallelExecutor) executeJob(job *ExecutionJob, worker *Worker, pool *ExecutionPool) {
	defer func() {
		pool.workers <- worker
		pool.wg.Done()
	}()

	pe.mu.Lock()
	job.AssignedWorker = worker
	job.Status = JobStatusRunning
	job.StartTime = time.Now()
	pe.activeJobs[job.ID] = job
	pe.mu.Unlock()

	worker.Status = WorkerStatusBusy
	worker.CurrentTask = &job.ID
	pool.activeWorkers.Store(worker.ID, worker)

	defer func() {
		worker.Status = WorkerStatusIdle
		worker.CurrentTask = nil
		worker.LastSeen = time.Now()
		pool.activeWorkers.Delete(worker.ID)

		pe.mu.Lock()
		delete(pe.activeJobs, job.ID)
		pe.mu.Unlock()
	}()

	if pe.eventBus != nil {
		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    pe.getTaskIDFromJob(job),
			Type:      TaskEventStarted,
			Timestamp: time.Now(),
			Data: map[string]any{
				"job_id":    job.ID,
				"job_type":  job.Type,
				"worker_id": worker.ID,
			},
		}
		pe.eventBus.Publish(job.Context, event)
	}

	result, err := pe.executeWithTimeout(job)

	now := time.Now()
	job.EndTime = &now
	job.Result = result

	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err
		pe.updateMetrics(false, time.Since(job.StartTime))
	} else {
		job.Status = JobStatusCompleted
		pe.updateMetrics(true, time.Since(job.StartTime))
	}

	if pe.eventBus != nil {
		eventType := TaskEventCompleted
		if err != nil {
			eventType = TaskEventFailed
		}

		event := TaskEvent{
			ID:        generateEventID(),
			TaskID:    pe.getTaskIDFromJob(job),
			Type:      eventType,
			Timestamp: time.Now(),
			Data: map[string]any{
				"job_id":   job.ID,
				"duration": time.Since(job.StartTime),
				"result":   result,
				"error":    err,
			},
		}
		pe.eventBus.Publish(job.Context, event)
	}
}

func (pe *ParallelExecutor) executeWithTimeout(job *ExecutionJob) (*ExecutionResult, error) {
	timeout := pe.config.DefaultJobTimeout
	
	jobCtx, cancel := context.WithTimeout(job.Context, timeout)
	defer cancel()

	job.Cancel = cancel

	resultChan := make(chan struct {
		result *ExecutionResult
		err    error
	}, 1)

	go func() {
		result, err := job.Executor(jobCtx, job)
		resultChan <- struct {
			result *ExecutionResult
			err    error
		}{result, err}
	}()

	select {
	case res := <-resultChan:
		return res.result, res.err
	case <-jobCtx.Done():
		return nil, fmt.Errorf("job execution timeout after %v", timeout)
	}
}

func (pe *ParallelExecutor) SubmitJob(ctx context.Context, job *ExecutionJob) error {
	if job.ID == "" {
		job.ID = generateJobID()
	}

	if job.Context == nil {
		job.Context = ctx
	}

	job.Status = JobStatusQueued

	pe.mu.Lock()
	pe.activeJobs[job.ID] = job
	pe.mu.Unlock()

	select {
	case pe.executionPool.jobQueue <- job:
		return nil
	default:
		pe.mu.Lock()
		delete(pe.activeJobs, job.ID)
		pe.mu.Unlock()
		return fmt.Errorf("job queue is full")
	}
}

func (pe *ParallelExecutor) SubmitTask(ctx context.Context, task *Task, executor JobExecutorFunc) (*ExecutionJob, error) {
	job := &ExecutionJob{
		ID:       generateJobID(),
		Type:     JobTypeTask,
		Task:     task,
		Context:  ctx,
		Status:   JobStatusPending,
		Executor: executor,
	}

	err := pe.SubmitJob(ctx, job)
	return job, err
}

func (pe *ParallelExecutor) SubmitSubTask(ctx context.Context, subtask *SubTask, executor JobExecutorFunc) (*ExecutionJob, error) {
	job := &ExecutionJob{
		ID:       generateJobID(),
		Type:     JobTypeSubTask,
		SubTask:  subtask,
		Context:  ctx,
		Status:   JobStatusPending,
		Executor: executor,
	}

	err := pe.SubmitJob(ctx, job)
	return job, err
}

func (pe *ParallelExecutor) SubmitStep(ctx context.Context, step *TaskStep, executor JobExecutorFunc) (*ExecutionJob, error) {
	job := &ExecutionJob{
		ID:       generateJobID(),
		Type:     JobTypeStep,
		Step:     step,
		Context:  ctx,
		Status:   JobStatusPending,
		Executor: executor,
	}

	err := pe.SubmitJob(ctx, job)
	return job, err
}

func (pe *ParallelExecutor) SubmitBatch(ctx context.Context, jobs []*ExecutionJob) error {
	for _, job := range jobs {
		if err := pe.SubmitJob(ctx, job); err != nil {
			return fmt.Errorf("failed to submit job %s: %w", job.ID, err)
		}
	}
	return nil
}

func (pe *ParallelExecutor) WaitForJob(ctx context.Context, jobID string) (*ExecutionResult, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			pe.mu.RLock()
			job, exists := pe.activeJobs[jobID]
			pe.mu.RUnlock()

			if !exists {
				return nil, fmt.Errorf("job not found: %s", jobID)
			}

			if job.Status == JobStatusCompleted || job.Status == JobStatusFailed || job.Status == JobStatusCancelled {
				if job.Error != nil {
					return job.Result, job.Error
				}
				return job.Result, nil
			}
		}
	}
}

func (pe *ParallelExecutor) WaitForJobs(ctx context.Context, jobIDs []string) ([]*ExecutionResult, error) {
	results := make([]*ExecutionResult, len(jobIDs))
	
	for i, jobID := range jobIDs {
		result, err := pe.WaitForJob(ctx, jobID)
		if err != nil {
			return nil, fmt.Errorf("job %s failed: %w", jobID, err)
		}
		results[i] = result
	}

	return results, nil
}

func (pe *ParallelExecutor) CancelJob(ctx context.Context, jobID string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	job, exists := pe.activeJobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Cancel != nil {
		job.Cancel()
	}

	job.Status = JobStatusCancelled
	pe.updateMetrics(false, time.Since(job.StartTime))

	return nil
}

func (pe *ParallelExecutor) GetJobStatus(ctx context.Context, jobID string) (*ExecutionJob, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	job, exists := pe.activeJobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	jobCopy := *job
	return &jobCopy, nil
}

func (pe *ParallelExecutor) ListActiveJobs(ctx context.Context) ([]*ExecutionJob, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	jobs := make([]*ExecutionJob, 0, len(pe.activeJobs))
	for _, job := range pe.activeJobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}

	return jobs, nil
}

func (pe *ParallelExecutor) GetMetrics(ctx context.Context) *ExecutorMetrics {
	pe.metrics.mu.RLock()
	defer pe.metrics.mu.RUnlock()

	metricsCopy := *pe.metrics
	return &metricsCopy
}

func (pe *ParallelExecutor) updateMetrics(success bool, duration time.Duration) {
	pe.metrics.mu.Lock()
	defer pe.metrics.mu.Unlock()

	pe.metrics.TotalJobsExecuted++
	if success {
		pe.metrics.SuccessfulJobs++
	} else {
		pe.metrics.FailedJobs++
	}

	totalDuration := time.Duration(pe.metrics.TotalJobsExecuted-1)*pe.metrics.AvgExecutionTime + duration
	pe.metrics.AvgExecutionTime = totalDuration / time.Duration(pe.metrics.TotalJobsExecuted)

	pe.metrics.CurrentConcurrentJobs = len(pe.activeJobs)
	if pe.metrics.CurrentConcurrentJobs > pe.metrics.PeakConcurrentJobs {
		pe.metrics.PeakConcurrentJobs = pe.metrics.CurrentConcurrentJobs
	}

	pe.metrics.LastUpdateTime = time.Now()
}

func (pe *ParallelExecutor) Shutdown(ctx context.Context) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	for _, job := range pe.activeJobs {
		if job.Cancel != nil {
			job.Cancel()
		}
	}

	close(pe.executionPool.shutdown)

	done := make(chan struct{})
	go func() {
		pe.executionPool.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (pe *ParallelExecutor) getTaskIDFromJob(job *ExecutionJob) string {
	switch job.Type {
	case JobTypeTask:
		if job.Task != nil {
			return job.Task.ID
		}
	case JobTypeSubTask:
		if job.SubTask != nil {
			return job.SubTask.ParentTaskID
		}
	case JobTypeStep:
		return ""
	}
	return ""
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}