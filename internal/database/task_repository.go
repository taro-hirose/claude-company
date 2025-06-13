package database

import (
	"database/sql"
	"time"

	"claude-company/internal/models"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		db: GetDB(),
	}
}

func (r *TaskRepository) Create(task *models.Task) error {
	query := `
		INSERT INTO tasks (id, parent_id, description, mode, pane_id, status, priority, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(query,
		task.ID,
		task.ParentID,
		task.Description,
		task.Mode,
		task.PaneID,
		task.Status,
		task.Priority,
		task.CreatedAt,
		task.UpdatedAt,
		task.Metadata,
	)
	return err
}

func (r *TaskRepository) GetByID(id string) (*models.Task, error) {
	query := `
		SELECT id, parent_id, description, mode, pane_id, status, priority,
		       created_at, updated_at, completed_at, result, metadata
		FROM tasks WHERE id = $1`

	task := &models.Task{}
	var parentID sql.NullString
	var completedAt sql.NullTime
	var result sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&parentID,
		&task.Description,
		&task.Mode,
		&task.PaneID,
		&task.Status,
		&task.Priority,
		&task.CreatedAt,
		&task.UpdatedAt,
		&completedAt,
		&result,
		&metadata,
	)

	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		task.ParentID = &parentID.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if result.Valid {
		task.Result = result.String
	}
	if metadata.Valid {
		task.Metadata = metadata.String
	}

	return task, nil
}

func (r *TaskRepository) GetByPaneID(paneID string) ([]*models.Task, error) {
	query := `
		SELECT id, parent_id, description, mode, pane_id, status, priority,
		       created_at, updated_at, completed_at, result, metadata
		FROM tasks 
		WHERE pane_id = $1 OR id IN (
			SELECT task_id FROM task_shares WHERE shared_with_pane_id = $1
		)
		ORDER BY created_at DESC`

	return r.queryTasks(query, paneID)
}

func (r *TaskRepository) GetChildren(parentID string) ([]*models.Task, error) {
	query := `
		SELECT id, parent_id, description, mode, pane_id, status, priority,
		       created_at, updated_at, completed_at, result, metadata
		FROM tasks WHERE parent_id = $1
		ORDER BY created_at ASC`

	return r.queryTasks(query, parentID)
}

func (r *TaskRepository) GetByStatus(status string) ([]*models.Task, error) {
	query := `
		SELECT id, parent_id, description, mode, pane_id, status, priority,
		       created_at, updated_at, completed_at, result, metadata
		FROM tasks WHERE status = $1
		ORDER BY priority DESC, created_at ASC`

	return r.queryTasks(query, status)
}

func (r *TaskRepository) Update(task *models.Task) error {
	query := `
		UPDATE tasks SET
			description = $2,
			mode = $3,
			pane_id = $4,
			status = $5,
			priority = $6,
			updated_at = $7,
			completed_at = $8,
			result = $9,
			metadata = $10
		WHERE id = $1`

	_, err := r.db.Exec(query,
		task.ID,
		task.Description,
		task.Mode,
		task.PaneID,
		task.Status,
		task.Priority,
		task.UpdatedAt,
		task.CompletedAt,
		task.Result,
		task.Metadata,
	)
	return err
}

func (r *TaskRepository) UpdateStatus(id, status string) error {
	var completedAt *time.Time
	if status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE tasks SET
			status = $2,
			updated_at = $3,
			completed_at = $4
		WHERE id = $1`

	_, err := r.db.Exec(query, id, status, time.Now(), completedAt)
	return err
}

func (r *TaskRepository) Delete(id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TaskRepository) DeleteByPaneID(paneID string) error {
	query := `DELETE FROM tasks WHERE pane_id = $1`
	_, err := r.db.Exec(query, paneID)
	return err
}

func (r *TaskRepository) queryTasks(query string, args ...interface{}) ([]*models.Task, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var parentID sql.NullString
		var completedAt sql.NullTime
		var result sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&task.ID,
			&parentID,
			&task.Description,
			&task.Mode,
			&task.PaneID,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
			&completedAt,
			&result,
			&metadata,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			task.ParentID = &parentID.String
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if result.Valid {
			task.Result = result.String
		}
		if metadata.Valid {
			task.Metadata = metadata.String
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

type TaskShare struct {
	ID              string    `json:"id" db:"id"`
	TaskID          string    `json:"task_id" db:"task_id"`
	SharedWithPaneID string   `json:"shared_with_pane_id" db:"shared_with_pane_id"`
	Permission      string    `json:"permission" db:"permission"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

func (r *TaskRepository) ShareTask(taskID, paneID, permission string) error {
	shareID := models.GenerateULID()
	query := `
		INSERT INTO task_shares (id, task_id, shared_with_pane_id, permission, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (task_id, shared_with_pane_id)
		DO UPDATE SET permission = $4, created_at = $5`

	_, err := r.db.Exec(query, shareID, taskID, paneID, permission, time.Now())
	return err
}

func (r *TaskRepository) UnshareTask(taskID, paneID string) error {
	query := `DELETE FROM task_shares WHERE task_id = $1 AND shared_with_pane_id = $2`
	_, err := r.db.Exec(query, taskID, paneID)
	return err
}

func (r *TaskRepository) GetTaskShares(taskID string) ([]*TaskShare, error) {
	query := `
		SELECT id, task_id, shared_with_pane_id, permission, created_at
		FROM task_shares WHERE task_id = $1`

	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []*TaskShare
	for rows.Next() {
		share := &TaskShare{}
		err := rows.Scan(
			&share.ID,
			&share.TaskID,
			&share.SharedWithPaneID,
			&share.Permission,
			&share.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	return shares, rows.Err()
}

func (r *TaskRepository) GetSharedTasks(paneID string) ([]*models.Task, error) {
	query := `
		SELECT t.id, t.parent_id, t.description, t.mode, t.pane_id, t.status, t.priority,
		       t.created_at, t.updated_at, t.completed_at, t.result, t.metadata
		FROM tasks t
		INNER JOIN task_shares ts ON t.id = ts.task_id
		WHERE ts.shared_with_pane_id = $1
		ORDER BY t.created_at DESC`

	return r.queryTasks(query, paneID)
}

// GetByPane はGetByPaneIDのエイリアス
func (r *TaskRepository) GetByPane(paneID string) ([]*models.Task, error) {
	return r.GetByPaneID(paneID)
}