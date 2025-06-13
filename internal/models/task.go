package models

type Task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Mode        string `json:"mode"`
	PaneID      string `json:"pane_id"`
	Status      string `json:"status"`
}