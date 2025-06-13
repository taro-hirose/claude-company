package models

import (
	"fmt"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusNeedsReview TaskStatus = "needs_review"
	TaskStatusRevisionRequired TaskStatus = "revision_required"
)

type SubTask struct {
	ID          string    `json:"id"`
	ParentTaskID string   `json:"parent_task_id"`
	Description string    `json:"description"`
	AssignedPane string   `json:"assigned_pane"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string    `json:"result,omitempty"`
	ReviewNotes string    `json:"review_notes,omitempty"`
}

type TaskTracker struct {
	MainTask Task         `json:"main_task"`
	SubTasks []SubTask    `json:"sub_tasks"`
	ManagerPane string    `json:"manager_pane"`
}

func NewTaskTracker(mainTask Task, managerPane string) *TaskTracker {
	return &TaskTracker{
		MainTask:    mainTask,
		SubTasks:    make([]SubTask, 0),
		ManagerPane: managerPane,
	}
}

func (t *TaskTracker) AddSubTask(description, assignedPane string) SubTask {
	subTask := SubTask{
		ID:           generateSubTaskID(t.MainTask.ID, len(t.SubTasks)+1),
		ParentTaskID: t.MainTask.ID,
		Description:  description,
		AssignedPane: assignedPane,
		Status:       TaskStatusPending,
		CreatedAt:    time.Now(),
	}
	t.SubTasks = append(t.SubTasks, subTask)
	return subTask
}

func (t *TaskTracker) UpdateSubTaskStatus(subTaskID string, status TaskStatus, result string) bool {
	for i, task := range t.SubTasks {
		if task.ID == subTaskID {
			t.SubTasks[i].Status = status
			if result != "" {
				t.SubTasks[i].Result = result
			}
			if status == TaskStatusCompleted || status == TaskStatusNeedsReview {
				now := time.Now()
				t.SubTasks[i].CompletedAt = &now
			}
			return true
		}
	}
	return false
}

func (t *TaskTracker) GetPendingTasks() []SubTask {
	var pending []SubTask
	for _, task := range t.SubTasks {
		if task.Status == TaskStatusPending || task.Status == TaskStatusRevisionRequired {
			pending = append(pending, task)
		}
	}
	return pending
}

func (t *TaskTracker) GetTasksNeedingReview() []SubTask {
	var needsReview []SubTask
	for _, task := range t.SubTasks {
		if task.Status == TaskStatusNeedsReview {
			needsReview = append(needsReview, task)
		}
	}
	return needsReview
}

func (t *TaskTracker) AllTasksCompleted() bool {
	for _, task := range t.SubTasks {
		if task.Status != TaskStatusCompleted {
			return false
		}
	}
	return len(t.SubTasks) > 0
}

func generateSubTaskID(parentID string, sequence int) string {
	return fmt.Sprintf("%s_sub_%d", parentID, sequence)
}