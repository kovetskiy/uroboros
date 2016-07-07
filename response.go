package main

type ResponseTaskQueued struct {
	ID int64 `json:"id"`
}

type ResponseTask struct {
	UniqueID    int64 `json:"unique_id"`
	Identifier string `json:"identifier"`
	State string   `json:"state"`
	Title string   `json:"title"`
	Logs  []string `json:"logs,omitempty"`
}

type ResponseTaskList struct {
	Tasks []ResponseTask `json:"tasks"`
}
