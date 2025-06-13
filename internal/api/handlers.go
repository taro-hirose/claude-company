package api

import (
	"net/http"

	"claude-company/internal/database"
	"claude-company/internal/models"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	repo    *database.TaskRepository
	service *TaskService
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		repo:    database.NewTaskRepository(),
		service: NewTaskService(nil), // sessionManagerは別途設定する必要がある
	}
}

type CreateTaskRequest struct {
	ParentID    *string `json:"parent_id,omitempty"`
	Description string  `json:"description" binding:"required"`
	Mode        string  `json:"mode" binding:"required"`
	PaneID      string  `json:"pane_id" binding:"required"`
	Priority    int     `json:"priority,omitempty"`
	Metadata    string  `json:"metadata,omitempty"`
}

type UpdateTaskRequest struct {
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	Result      string `json:"result,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
}

type ShareTaskRequest struct {
	PaneID     string `json:"pane_id" binding:"required"`
	Permission string `json:"permission,omitempty"`
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var task *models.Task
	if req.ParentID != nil {
		task = models.NewSubTask(*req.ParentID, req.Description, req.Mode, req.PaneID)
	} else {
		task = models.NewTask(req.Description, req.Mode, req.PaneID)
	}

	if req.Priority > 0 {
		task.Priority = req.Priority
	}
	if req.Metadata != "" {
		task.Metadata = req.Metadata
	}

	if err := h.repo.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	paneID := c.Query("pane_id")
	status := c.Query("status")
	parentID := c.Query("parent_id")

	var tasks []*models.Task
	var err error

	switch {
	case parentID != "":
		tasks, err = h.repo.GetChildren(parentID)
	case status != "":
		tasks, err = h.repo.GetByStatus(status)
	case paneID != "":
		tasks, err = h.repo.GetByPaneID(paneID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "pane_id, status, or parent_id parameter required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	
	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.UpdateStatus(req.Status)
		if req.Status == "completed" && req.Result != "" {
			task.MarkCompleted(req.Result)
		}
	}
	if req.Priority > 0 {
		task.Priority = req.Priority
	}
	if req.Result != "" {
		task.Result = req.Result
	}
	if req.Metadata != "" {
		task.Metadata = req.Metadata
	}

	if err := h.repo.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Param("status")

	if err := h.repo.UpdateStatus(id, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task status"})
		return
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func (h *TaskHandler) ShareTask(c *gin.Context) {
	id := c.Param("id")
	
	var req ShareTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission := req.Permission
	if permission == "" {
		permission = "read"
	}

	if err := h.repo.ShareTask(id, req.PaneID, permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task shared successfully"})
}

func (h *TaskHandler) UnshareTask(c *gin.Context) {
	id := c.Param("id")
	paneID := c.Param("pane_id")

	if err := h.repo.UnshareTask(id, paneID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unshare task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task unshared successfully"})
}

func (h *TaskHandler) GetTaskShares(c *gin.Context) {
	id := c.Param("id")

	shares, err := h.repo.GetTaskShares(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task shares"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shares": shares})
}

func (h *TaskHandler) GetSharedTasks(c *gin.Context) {
	paneID := c.Query("pane_id")
	if paneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pane_id parameter required"})
		return
	}

	tasks, err := h.repo.GetSharedTasks(paneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shared tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

type ProgressResponse struct {
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	PendingTasks    int     `json:"pending_tasks"`
	InProgressTasks int     `json:"in_progress_tasks"`
	ProgressPercent float64 `json:"progress_percent"`
}

func (h *TaskHandler) GetProgress(c *gin.Context) {
	paneID := c.Query("pane_id")
	if paneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pane_id parameter required"})
		return
	}

	tasks, err := h.repo.GetByPaneID(paneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	progress := ProgressResponse{
		TotalTasks: len(tasks),
	}

	for _, task := range tasks {
		switch task.Status {
		case "completed":
			progress.CompletedTasks++
		case "in_progress":
			progress.InProgressTasks++
		case "pending":
			progress.PendingTasks++
		}
	}

	if progress.TotalTasks > 0 {
		progress.ProgressPercent = float64(progress.CompletedTasks) / float64(progress.TotalTasks) * 100
	}

	c.JSON(http.StatusOK, progress)
}

func (h *TaskHandler) GetTaskHierarchy(c *gin.Context) {
	id := c.Param("id")
	
	hierarchy, err := h.service.GetTaskHierarchy(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found or failed to fetch hierarchy"})
		return
	}

	c.JSON(http.StatusOK, hierarchy)
}

func (h *TaskHandler) ShareWithSiblings(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.service.ShareTaskWithSiblings(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task shared with siblings successfully"})
}

func (h *TaskHandler) ShareWithFamily(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.service.ShareTaskWithFamily(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task shared with family successfully"})
}

func (h *TaskHandler) UpdateTaskStatusWithPropagation(c *gin.Context) {
	id := c.Param("id")
	status := c.Param("status")

	if err := h.service.PropagateStatusUpdate(id, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task status"})
		return
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetTaskStatistics(c *gin.Context) {
	paneID := c.Query("pane_id")
	if paneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pane_id parameter required"})
		return
	}

	stats, err := h.service.GetTaskStatistics(paneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}