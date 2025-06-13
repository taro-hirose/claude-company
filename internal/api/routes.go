package api

import (
	"claude-company/internal/database"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(config *database.Config) *gin.Engine {
	if err := database.InitDB(config); err != nil {
		panic("Failed to initialize database: " + err.Error())
	}

	r := gin.Default()
	
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	taskHandler := NewTaskHandler()

	api := r.Group("/api/v1")
	{
		tasks := api.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.GetTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.GET("/:id/hierarchy", taskHandler.GetTaskHierarchy)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.PATCH("/:id/status/:status", taskHandler.UpdateTaskStatus)
			tasks.PATCH("/:id/status-propagate/:status", taskHandler.UpdateTaskStatusWithPropagation)
			tasks.DELETE("/:id", taskHandler.DeleteTask)

			tasks.POST("/:id/share", taskHandler.ShareTask)
			tasks.POST("/:id/share-siblings", taskHandler.ShareWithSiblings)
			tasks.POST("/:id/share-family", taskHandler.ShareWithFamily)
			tasks.DELETE("/:id/share/:pane_id", taskHandler.UnshareTask)
			tasks.GET("/:id/shares", taskHandler.GetTaskShares)
		}

		api.GET("/shared-tasks", taskHandler.GetSharedTasks)
		api.GET("/progress", taskHandler.GetProgress)
		api.GET("/statistics", taskHandler.GetTaskStatistics)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Claude Company API is running",
		})
	})

	return r
}