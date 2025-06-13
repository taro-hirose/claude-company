-- Claude Company Database Schema
-- Tasks and Task Progress Tables with ULID, parent-child relationships, and progress tracking

-- Extension for ULID generation (if not available, can use UUID instead)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Task status enum
CREATE TYPE task_status AS ENUM (
    'pending',
    'in_progress',
    'completed',
    'cancelled',
    'blocked'
);

-- Task priority enum
CREATE TYPE task_priority AS ENUM (
    'low',
    'medium',
    'high',
    'urgent'
);

-- Tasks table with hierarchical structure
CREATE TABLE tasks (
    id VARCHAR(26) PRIMARY KEY, -- ULID format
    parent_id VARCHAR(26) REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status task_status NOT NULL DEFAULT 'pending',
    priority task_priority NOT NULL DEFAULT 'medium',
    assigned_to VARCHAR(100), -- User/agent identifier
    created_by VARCHAR(100) NOT NULL,
    estimated_duration_minutes INTEGER,
    due_date TIMESTAMP WITH TIME ZONE,
    tags TEXT[], -- Array of tags for categorization
    metadata JSONB, -- Flexible metadata storage
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Task progress tracking table
CREATE TABLE task_progress (
    id VARCHAR(26) PRIMARY KEY, -- ULID format
    task_id VARCHAR(26) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    progress_percentage INTEGER NOT NULL DEFAULT 0 CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    status task_status NOT NULL,
    notes TEXT,
    time_spent_minutes INTEGER DEFAULT 0,
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_created_by ON tasks(created_by);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_tags ON tasks USING GIN(tags);

CREATE INDEX idx_task_progress_task_id ON task_progress(task_id);
CREATE INDEX idx_task_progress_created_at ON task_progress(created_at);
CREATE INDEX idx_task_progress_status ON task_progress(status);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at
CREATE TRIGGER update_tasks_updated_at 
    BEFORE UPDATE ON tasks 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to get task hierarchy depth
CREATE OR REPLACE FUNCTION get_task_depth(task_id VARCHAR(26))
RETURNS INTEGER AS $$
WITH RECURSIVE task_hierarchy AS (
    SELECT id, parent_id, 0 as depth
    FROM tasks
    WHERE id = task_id
    
    UNION ALL
    
    SELECT t.id, t.parent_id, th.depth + 1
    FROM tasks t
    INNER JOIN task_hierarchy th ON t.id = th.parent_id
)
SELECT MAX(depth) FROM task_hierarchy;
$$ LANGUAGE SQL;

-- Function to get all subtasks
CREATE OR REPLACE FUNCTION get_subtasks(task_id VARCHAR(26))
RETURNS TABLE(
    id VARCHAR(26),
    title VARCHAR(255),
    status task_status,
    priority task_priority,
    depth INTEGER
) AS $$
WITH RECURSIVE task_tree AS (
    SELECT t.id, t.title, t.status, t.priority, t.parent_id, 0 as depth
    FROM tasks t
    WHERE t.id = task_id
    
    UNION ALL
    
    SELECT t.id, t.title, t.status, t.priority, t.parent_id, tt.depth + 1
    FROM tasks t
    INNER JOIN task_tree tt ON t.parent_id = tt.id
)
SELECT tt.id, tt.title, tt.status, tt.priority, tt.depth
FROM task_tree tt
WHERE tt.depth > 0
ORDER BY tt.depth, tt.title;
$$ LANGUAGE SQL;

-- View for task summary with progress
CREATE VIEW task_summary AS
SELECT 
    t.id,
    t.parent_id,
    t.title,
    t.description,
    t.status,
    t.priority,
    t.assigned_to,
    t.created_by,
    t.estimated_duration_minutes,
    t.due_date,
    t.tags,
    t.created_at,
    t.updated_at,
    COALESCE(tp.latest_progress, 0) as current_progress,
    COALESCE(tp.total_time_spent, 0) as total_time_spent_minutes,
    (SELECT COUNT(*) FROM tasks subtasks WHERE subtasks.parent_id = t.id) as subtask_count
FROM tasks t
LEFT JOIN (
    SELECT 
        task_id,
        MAX(progress_percentage) as latest_progress,
        SUM(time_spent_minutes) as total_time_spent
    FROM task_progress
    GROUP BY task_id
) tp ON t.id = tp.task_id;