package utils

import (
	"fmt"
	"os"
)

func GenerateTaskID() string {
	return fmt.Sprintf("task_%d", os.Getpid())
}