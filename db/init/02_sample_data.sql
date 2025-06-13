-- Sample data for testing
-- Note: In production, you would generate proper ULIDs

INSERT INTO tasks (id, title, description, status, priority, assigned_to, created_by, estimated_duration_minutes, tags) VALUES
('01HK8X7D9M0000000000000001', 'Project Setup', 'Initialize the Claude Company project', 'completed', 'high', 'claude-agent', 'system', 120, ARRAY['setup', 'initialization']),
('01HK8X7D9M0000000000000002', 'Database Design', 'Design and implement database schema', 'in_progress', 'high', 'claude-agent', 'system', 180, ARRAY['database', 'schema']),
('01HK8X7D9M0000000000000003', 'API Development', 'Develop REST API endpoints', 'pending', 'medium', 'claude-agent', 'system', 300, ARRAY['api', 'backend']),
('01HK8X7D9M0000000000000004', 'Task Management', 'Implement task management features', 'pending', 'high', 'claude-agent', 'system', 240, ARRAY['tasks', 'management']),
('01HK8X7D9M0000000000000005', 'Create Tables', 'Create database tables and indexes', 'in_progress', 'high', 'claude-agent', 'system', 60, ARRAY['database', 'tables']);

-- Set parent-child relationships
UPDATE tasks SET parent_id = '01HK8X7D9M0000000000000002' WHERE id = '01HK8X7D9M0000000000000005';

-- Insert progress tracking data
INSERT INTO task_progress (id, task_id, progress_percentage, status, notes, time_spent_minutes, created_by) VALUES
('01HK8X7D9N0000000000000001', '01HK8X7D9M0000000000000001', 100, 'completed', 'Project structure created successfully', 120, 'claude-agent'),
('01HK8X7D9N0000000000000002', '01HK8X7D9M0000000000000002', 75, 'in_progress', 'Schema design completed, implementing tables', 135, 'claude-agent'),
('01HK8X7D9N0000000000000003', '01HK8X7D9M0000000000000005', 80, 'in_progress', 'Main tables created, working on indexes', 48, 'claude-agent');