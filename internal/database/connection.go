package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "claude_user",
		Password: "claude_password",
		DBName:   "claude_company",
		SSLMode:  "disable",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func InitDB(config *Config) error {
	var err error
	once.Do(func() {
		db, err = sql.Open("postgres", config.ConnectionString())
		if err != nil {
			return
		}

		if err = db.Ping(); err != nil {
			return
		}

		if err = createTables(); err != nil {
			return
		}

		log.Println("Database connection established successfully")
	})
	return err
}

func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("Database not initialized. Call InitDB first.")
	}
	return db
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func createTables() error {
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(26) PRIMARY KEY,
		parent_id VARCHAR(26) REFERENCES tasks(id) ON DELETE CASCADE,
		description TEXT NOT NULL,
		mode VARCHAR(50) NOT NULL,
		pane_id VARCHAR(100) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		priority INTEGER NOT NULL DEFAULT 1,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		completed_at TIMESTAMP WITH TIME ZONE,
		result TEXT,
		metadata JSONB
	);`

	createTaskSharesTable := `
	CREATE TABLE IF NOT EXISTS task_shares (
		id VARCHAR(26) PRIMARY KEY,
		task_id VARCHAR(26) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		shared_with_pane_id VARCHAR(100) NOT NULL,
		permission VARCHAR(20) NOT NULL DEFAULT 'read',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		UNIQUE(task_id, shared_with_pane_id)
	);`

	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_pane_id ON tasks(pane_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_task_shares_task_id ON task_shares(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_shares_pane_id ON task_shares(shared_with_pane_id);`

	for _, query := range []string{createTasksTable, createTaskSharesTable, createIndexes} {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create tables: %w", err)
		}
	}

	return nil
}